package forward

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/expr-lang/expr/vm"
	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/telegram/message/html"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/atomic"
	"go.uber.org/zap"

	"github.com/iyear/tdl/core/dcpool"
	"github.com/iyear/tdl/core/forwarder"
	"github.com/iyear/tdl/core/logctx"
	"github.com/iyear/tdl/core/util/tutil"
	"github.com/iyear/tdl/pkg/texpr"
	"github.com/iyear/tdl/pkg/tmessage"
)

type iterOptions struct {
	manager  *peers.Manager
	pool     dcpool.Pool
	to       *vm.Program
	edit     *vm.Program
	dialogs  []*tmessage.Dialog
	mode     forwarder.Mode
	silent   bool
	dryRun   bool
	grouped  bool
	delay    time.Duration
	skipSame bool
}

type iter struct {
	opts iterOptions

	seen    map[string]struct{} // forwarded media keys, used by skip-same
	skipped *atomic.Int64       // count of messages skipped by skip-same

	i, j int
	elem forwarder.Elem
	err  error
}

type env struct {
	From struct {
		ID          int64  `comment:"ID of dialog"`
		Username    string `comment:"Username of dialog"`
		VisibleName string `comment:"Title of channel and group, first and last name of user"`
	}
	Message texpr.EnvMessage
}

func exprEnv(from peers.Peer, msg *tg.Message) env {
	e := env{}

	if from != nil {
		e.From.ID = from.ID()
		e.From.Username, _ = from.Username()
		e.From.VisibleName = from.VisibleName()
	}

	if msg != nil {
		e.Message = texpr.ConvertEnvMessage(msg)
	}

	return e
}

type dest struct {
	Peer   string
	Thread int
}

func newIter(opts iterOptions) *iter {
	return &iter{
		opts: opts,

		seen:    make(map[string]struct{}),
		skipped: atomic.NewInt64(0),

		i:    0,
		j:    0,
		elem: nil,
		err:  nil,
	}
}

func (i *iter) Next(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		i.err = ctx.Err()
		return false
	default:
	}

	// end of iteration or error occurred
	if i.i >= len(i.opts.dialogs) || i.err != nil {
		return false
	}

	// if delay is set, sleep for a while for each iteration
	if i.opts.delay > 0 && (i.i+i.j) > 0 { // skip first delay
		time.Sleep(i.opts.delay)
	}

	for {
		// end of iteration
		if i.i >= len(i.opts.dialogs) {
			return false
		}

		p, m := i.opts.dialogs[i.i].Peer, i.opts.dialogs[i.i].Messages[i.j]

		if i.j++; i.j >= len(i.opts.dialogs[i.i].Messages) {
			i.i++
			i.j = 0
		}

		from, err := i.opts.manager.FromInputPeer(ctx, p)
		if err != nil {
			i.err = errors.Wrap(err, "get from peer")
			return false
		}

		msg, err := tutil.GetSingleMessage(ctx, i.opts.pool.Default(ctx), from.InputPeer(), m)
		if err != nil {
			i.err = errors.Wrapf(err, "get message: %d", m)
			return false
		}

		// message routing
		result, err := texpr.Run(i.opts.to, exprEnv(from, msg))
		if err != nil {
			i.err = errors.Wrap(err, "message routing")
			return false
		}

		var (
			to     peers.Peer
			thread int
		)

		switch r := result.(type) {
		case string:
			// pure chat, no reply to, which is a compatible with old version
			// and a convenient way to send message to self
			to, err = i.resolvePeer(ctx, r)
		case map[string]interface{}:
			// chat with reply to topic or message
			var d dest

			if err = mapstructure.WeakDecode(r, &d); err != nil {
				i.err = errors.Wrapf(err, "decode dest: %v", result)
				return false
			}

			to, err = i.resolvePeer(ctx, d.Peer)
			thread = d.Thread
		default:
			i.err = errors.Errorf("message router must return string or dest: %T", result)
			return false
		}

		if err != nil {
			i.err = errors.Wrapf(err, "resolve dest: %v", result)
			return false
		}

		// skip messages with the same file which have been forwarded
		// in this run to the same destination.
		//
		// albums are deduplicated as a whole by their primary message media:
		// children are not individually fingerprinted because an album is
		// always sent as one unit, and skipping it entirely on any child match
		// would drop unique files. dedup also applies to dry-run, so that the
		// preview matches what a real run would send.
		if i.opts.skipSame {
			if key, ok := mediaKey(msg); ok {
				k := mediaDestKey(to.ID(), thread, key)
				if _, dup := i.seen[k]; dup {
					i.skipped.Inc()
					logctx.From(ctx).Debug("skip duplicated message",
						zap.Int64("from", from.ID()),
						zap.Int64("to", to.ID()),
						zap.Int("message", msg.ID))
					continue
				}
				i.seen[k] = struct{}{}
			}
		}

		var modeOverride forwarder.Mode = -1 // default value is invalid
		// edit message
		if i.opts.edit != nil {
			result, err = texpr.Run(i.opts.edit, exprEnv(from, msg))
			if err != nil {
				i.err = errors.Wrap(err, "edit message")
				return false
			}

			r, ok := result.(string)
			if !ok {
				i.err = errors.Errorf("edit must return string: %T", result)
				return false
			}

			eb := entity.Builder{}
			if err = html.HTML(strings.NewReader(r), &eb, html.Options{
				UserResolver:          nil,
				DisableTelegramEscape: false,
			}); err != nil {
				i.err = errors.Wrap(err, "parse edited message")
				return false
			}

			// modify message
			msg.Message, msg.Entities = eb.Complete()
			// direct mode can't modify message content, so we force it to be clone mode
			modeOverride = forwarder.ModeClone
		}

		i.elem = &iterElem{
			from:         from,
			msg:          msg,
			to:           to,
			thread:       thread,
			modeOverride: modeOverride,
			opts:         i.opts,
		}

		return true
	}
}

func (i *iter) resolvePeer(ctx context.Context, peer string) (peers.Peer, error) {
	if peer == "" { // self
		return i.opts.manager.Self(ctx)
	}

	return tutil.GetInputPeer(ctx, i.opts.manager, peer)
}

// Skipped returns the number of messages skipped by the skip-same option.
func (i *iter) Skipped() int64 {
	return i.skipped.Load()
}

// mediaKey returns a stable identifier of the message media. It is used by
// the skip-same option to detect duplicate files. The second return value is
// false for messages without file media, which are never deduplicated.
func mediaKey(msg *tg.Message) (string, bool) {
	switch m := msg.Media.(type) {
	case *tg.MessageMediaDocument:
		doc, ok := m.Document.(*tg.Document)
		if !ok {
			return "", false
		}
		return fmt.Sprintf("doc:%d:%d", doc.ID, doc.Size), true
	case *tg.MessageMediaPhoto:
		photo, ok := m.Photo.(*tg.Photo)
		if !ok {
			return "", false
		}
		return fmt.Sprintf("photo:%d", photo.ID), true
	default:
		return "", false
	}
}

// mediaDestKey scopes a media key to a destination peer id and thread, so that
// forwarding the same file to different destinations is not skipped.
func mediaDestKey(toID int64, thread int, key string) string {
	return fmt.Sprintf("%d:%d:%s", toID, thread, key)
}

func (i *iter) Value() forwarder.Elem {
	return i.elem
}

func (i *iter) Err() error {
	return i.err
}
