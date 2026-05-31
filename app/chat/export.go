package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
	"github.com/fatih/color"
	"github.com/go-faster/jx"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/telegram/query/messages"
	"github.com/gotd/td/tg"
	"github.com/jedib0t/go-pretty/v6/progress"
	"go.uber.org/multierr"

	"github.com/iyear/tdl/core/storage"
	"github.com/iyear/tdl/core/tmedia"
	"github.com/iyear/tdl/core/util/tutil"
	"github.com/iyear/tdl/pkg/prog"
	"github.com/iyear/tdl/pkg/texpr"
)

//go:generate go-enum --names --values --flag --nocase

type ExportOptions struct {
	Type        ExportType
	Chat        string
	Thread      int // topic id in forum, message id in group
	Input       []int
	Output      string
	Filter      string
	OnlyMedia   bool
	WithContent bool
	Raw         bool
	All         bool
}

type Message struct {
	ID   int         `json:"id"`
	Type string      `json:"type"`
	File string      `json:"file"`
	Date int         `json:"date,omitempty"`
	Text string      `json:"text,omitempty"`
	Raw  *tg.Message `json:"raw,omitempty"`
}

// ExportType
// ENUM(time, id, last)
type ExportType int

func Export(ctx context.Context, c *telegram.Client, kvd storage.Storage, opts ExportOptions) (rerr error) {
	// only output available fields
	if opts.Filter == "-" {
		fg := texpr.NewFieldsGetter(nil)

		fields, err := fg.Walk(&texpr.EnvMessage{})
		if err != nil {
			return fmt.Errorf("failed to walk fields: %w", err)
		}

		fmt.Print(fg.Sprint(fields, true))
		return nil
	}

	filter, err := expr.Compile(opts.Filter, expr.AsBool())
	if err != nil {
		return fmt.Errorf("failed to compile filter: %w", err)
	}

	var peer peers.Peer

	manager := peers.Options{Storage: storage.NewPeers(kvd)}.Build(c.API())
	if opts.Chat == "" { // defaults to me(saved messages)
		peer, err = manager.Self(ctx)
	} else {
		peer, err = tutil.GetInputPeer(ctx, manager, opts.Chat)
	}
	if err != nil {
		return fmt.Errorf("failed to get peer: %w", err)
	}

	color.Yellow("WARN: Export only generates minimal JSON for tdl download, not for backup.")
	color.Cyan("Occasional suspensions are due to Telegram rate limitations, please wait a moment.")
	fmt.Println()

	color.Blue("Type: %s | Input: %v", opts.Type, opts.Input)

	pw := prog.New(progress.FormatNumber)
	pw.SetUpdateFrequency(200 * time.Millisecond)
	pw.Style().Visibility.TrackerOverall = false
	pw.Style().Visibility.ETA = false
	pw.Style().Visibility.Percentage = false

	tracker := prog.AppendTracker(pw, progress.FormatNumber, fmt.Sprintf("%s-%d", peer.VisibleName(), peer.ID()), 0)

	go pw.Render()

	var q messages.Query
	switch {
	case opts.Thread != 0: // topic messages, reply messages
		q = query.NewQuery(c.API()).Messages().GetReplies(peer.InputPeer()).MsgID(opts.Thread)
	default: // history
		q = query.NewQuery(c.API()).Messages().GetHistory(peer.InputPeer())
	}
	iter := messages.NewIterator(q, 100)

	switch opts.Type {
	case ExportTypeTime:
		iter = iter.OffsetDate(opts.Input[1] + 1)
	case ExportTypeId:
		iter = iter.OffsetID(opts.Input[1] + 1) // #89: retain the last msg id
	case ExportTypeLast:
	}

	f, err := os.Create(opts.Output)
	if err != nil {
		return err
	}
	defer multierr.AppendInvoke(&rerr, multierr.Close(f))

	enc := jx.NewStreamingEncoder(f, 512)
	defer multierr.AppendInvoke(&rerr, multierr.Close(enc))

	// process thread is reply type and peer is broadcast channel,
	// so we need to set discussion group id instead of broadcast id
	id := peer.ID()
	if p, ok := peer.(peers.Channel); opts.Thread != 0 && ok && p.IsBroadcast() {
		bc, _ := p.ToBroadcast()
		raw, err := bc.FullRaw(ctx)
		if err != nil {
			return fmt.Errorf("failed to get broadcast full raw: %w", err)
		}

		if id, ok = raw.GetLinkedChatID(); !ok {
			return fmt.Errorf("no linked group")
		}
	}

	enc.ObjStart()
	defer enc.ObjEnd()
	enc.Field("id", func(e *jx.Encoder) { e.Int64(id) })

	enc.FieldStart("messages")
	enc.ArrStart()
	defer enc.ArrEnd()

	count := int64(0)
	expander := newExportMessageExpander(opts, filter, func(ctx context.Context, msg *tg.Message) ([]*tg.Message, error) {
		return tutil.GetGroupedMessages(ctx, c.API(), peer.InputPeer(), msg)
	})

loop:
	for iter.Next(ctx) {
		msg := iter.Value()
		switch opts.Type {
		case ExportTypeTime:
			if msg.Msg.GetDate() < opts.Input[0] {
				break loop
			}
		case ExportTypeId:
			if msg.Msg.GetID() < opts.Input[0] {
				break loop
			}
		case ExportTypeLast:
			if count >= int64(opts.Input[0]) {
				break loop
			}
		}

		m, ok := msg.Msg.(*tg.Message)
		if !ok {
			continue
		}

		messages, err := expander.Expand(ctx, m)
		if err != nil {
			return err
		}
		for _, message := range messages {
			mb, err := json.Marshal(message)
			if err != nil {
				return fmt.Errorf("failed to marshal message: %w", err)
			}
			enc.Raw(mb)

			count++
			tracker.SetValue(count)
		}
	}

	if err = iter.Err(); err != nil {
		return err
	}

	tracker.MarkAsDone()
	prog.Wait(ctx, pw)
	return nil
}

type groupedMessageResolver func(context.Context, *tg.Message) ([]*tg.Message, error)

type exportMessageExpander struct {
	opts           ExportOptions
	filter         *vm.Program
	resolveGrouped groupedMessageResolver
	seen           map[int]struct{}
}

func newExportMessageExpander(
	opts ExportOptions,
	filter *vm.Program,
	resolveGrouped groupedMessageResolver,
) *exportMessageExpander {
	return &exportMessageExpander{
		opts:           opts,
		filter:         filter,
		resolveGrouped: resolveGrouped,
		seen:           make(map[int]struct{}),
	}
}

func (e *exportMessageExpander) Expand(ctx context.Context, msg *tg.Message) ([]*Message, error) {
	if _, ok := e.seen[msg.ID]; ok {
		return nil, nil
	}

	matched, err := e.matchesFilter(msg)
	if err != nil {
		return nil, err
	}
	if !matched {
		return nil, nil
	}

	messages := []*tg.Message{msg}
	if _, ok := msg.GetGroupedID(); ok && e.resolveGrouped != nil {
		grouped, err := e.resolveGrouped(ctx, msg)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve grouped message %d: %w", msg.ID, err)
		}
		if len(grouped) > 0 {
			messages = grouped
		}
	}

	exported := make([]*Message, 0, len(messages))
	for _, message := range messages {
		if _, ok := e.seen[message.ID]; ok {
			continue
		}

		out, ok := e.convert(message)
		if !ok {
			continue
		}

		e.seen[message.ID] = struct{}{}
		exported = append(exported, out)
	}

	return exported, nil
}

func (e *exportMessageExpander) matchesFilter(msg *tg.Message) (bool, error) {
	b, err := texpr.Run(e.filter, texpr.ConvertEnvMessage(msg))
	if err != nil {
		return false, fmt.Errorf("failed to run filter: %w", err)
	}

	matched, ok := b.(bool)
	if !ok {
		return false, fmt.Errorf("filter returned %T, expected bool", b)
	}
	return matched, nil
}

func (e *exportMessageExpander) convert(msg *tg.Message) (*Message, bool) {
	media, ok := tmedia.GetMedia(msg)
	if !ok && !e.opts.All {
		return nil, false
	}

	fileName := ""
	if media != nil { // #207
		fileName = media.Name
	}
	out := &Message{
		ID:   msg.ID,
		Type: "message",
		File: fileName,
	}
	if e.opts.WithContent {
		out.Date = msg.Date
		out.Text = msg.Message
	}
	if e.opts.Raw {
		out.Raw = msg
	}

	return out, true
}
