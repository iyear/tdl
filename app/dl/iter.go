package dl

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"text/template"
	"time"

	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
	"go.uber.org/atomic"
	"go.uber.org/zap"

	"github.com/iyear/tdl/core/dcpool"
	"github.com/iyear/tdl/core/downloader"
	"github.com/iyear/tdl/core/logctx"
	"github.com/iyear/tdl/core/tmedia"
	"github.com/iyear/tdl/core/util/fsutil"
	"github.com/iyear/tdl/core/util/tutil"
	"github.com/iyear/tdl/pkg/filterMap"
	"github.com/iyear/tdl/pkg/tmessage"
	"github.com/iyear/tdl/pkg/tplfunc"
	"github.com/iyear/tdl/pkg/utils"
)

const tempExt = ".tmp"

type fileTemplate struct {
	DialogID     int64
	MessageID    int
	MessageDate  int64
	FileName     string
	FileCaption  string
	FileSize     string
	DownloadDate int64
}

type iter struct {
	pool    dcpool.Pool
	manager *peers.Manager
	dialogs []*tmessage.Dialog
	tpl     *template.Template
	include map[string]struct{}
	exclude map[string]struct{}
	opts    Options
	delay   time.Duration

	mu          *sync.Mutex
	finished    map[int]struct{}
	fingerprint string
	// This param is kept for potential future use but is currently unused.
	// preSum       []int
	logicalPos   int // logical position for finished tracking
	dialogIndex  int // physical position: current dialog in dialogs array
	messageIndex int // physical position: current message in dialog.Messages array

	// TODO(Hexa): counter is de facto not be used in the codebase, but I perfer to reserve it. The key point is whether it still needs to be atomic or not.
	counter        *atomic.Int64
	skippedDeleted *atomic.Int64 // count of skipped deleted messages
	deletedIDs     []string      // IDs of deleted messages (format: "dialogID/messageID")
	elem           chan downloader.Elem
	err            error
}

func newIter(pool dcpool.Pool, manager *peers.Manager, dialog [][]*tmessage.Dialog,
	opts Options, delay time.Duration,
) (*iter, error) {
	tpl, err := template.New("dl").
		Funcs(tplfunc.FuncMap(tplfunc.All...)).
		Parse(opts.Template)
	if err != nil {
		return nil, errors.Wrap(err, "parse template")
	}

	dialogs := flatDialogs(dialog)
	// if msgs is empty, return error to avoid range out of index
	if len(dialogs) == 0 {
		return nil, errors.Errorf("you must specify at least one message")
	}

	// include and exclude
	includeMap := filterMap.New(opts.Include, fsutil.AddPrefixDot)
	excludeMap := filterMap.New(opts.Exclude, fsutil.AddPrefixDot)

	// to keep fingerprint stable
	sortDialogs(dialogs, opts.Desc)

	return &iter{
		pool:    pool,
		manager: manager,
		dialogs: dialogs,
		opts:    opts,
		include: includeMap,
		exclude: excludeMap,
		tpl:     tpl,
		delay:   delay,

		mu:          &sync.Mutex{},
		finished:    make(map[int]struct{}),
		fingerprint: fingerprint(dialogs),
		// This param is kept for potential future use but is currently unused.
		// preSum:       preSum(dialogs),
		logicalPos:     0,
		dialogIndex:    0,
		messageIndex:   0,
		counter:        atomic.NewInt64(-1),
		skippedDeleted: atomic.NewInt64(0),
		deletedIDs:     make([]string, 0),
		elem:           make(chan downloader.Elem, 10), // grouped message buffer
		err:            nil,
	}, nil
}

func (i *iter) Next(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		i.err = ctx.Err()
		return false
	default:
	}

	// if delay is set, sleep for a while for each iteration
	if i.delay > 0 && (i.dialogIndex+i.messageIndex) > 0 { // skip first delay
		time.Sleep(i.delay)
	}

	if len(i.elem) > 0 { // there are messages(grouped) in channel that not processed
		return true
	}

	for {
		ok, skip := i.process(ctx)
		if skip {
			continue
		}

		return ok
	}
}

func (i *iter) process(ctx context.Context) (ret bool, skip bool) {
	i.mu.Lock()
	defer i.mu.Unlock()

	// end of iteration or error occurred
	if i.dialogIndex >= len(i.dialogs) || i.messageIndex >= len(i.dialogs[i.dialogIndex].Messages) || i.err != nil {
		return false, false
	}

	peer, msg := i.dialogs[i.dialogIndex].Peer, i.dialogs[i.dialogIndex].Messages[i.messageIndex]

	// Record current logical position before processing
	startLogicalPos := i.logicalPos

	// Defer physical position increment
	defer func() {
		if i.messageIndex++; i.dialogIndex < len(i.dialogs) && i.messageIndex >= len(i.dialogs[i.dialogIndex].Messages) {
			i.dialogIndex++
			i.messageIndex = 0
		}
	}()

	from, err := i.manager.FromInputPeer(ctx, peer)
	if err != nil {
		i.err = errors.Wrap(err, "resolve from input peer")
		return false, false
	}
	message, err := tutil.GetSingleMessage(ctx, i.pool.Default(ctx), peer, msg)
	if err != nil {
		// Check if the error is due to a deleted message
		if errors.Is(err, tutil.ErrMessageDeleted) {
			logctx.From(ctx).Info("Message may be deleted, skipping",
				zap.Int64("dialog_id", tutil.GetInputPeerID(peer)),
				zap.Int("message_id", msg),
			)
			i.skippedDeleted.Inc()                                                                     // increment skipped deleted counter
			i.deletedIDs = append(i.deletedIDs, fmt.Sprintf("%d/%d", tutil.GetInputPeerID(peer), msg)) // track deleted message ID
			i.logicalPos++                                                                             // increment logical position for skipped message
			return false, true
		}
		i.err = errors.Wrap(err, "resolve message")
		return false, false
	}

	if _, ok := message.GetGroupedID(); ok && i.opts.Group {
		return i.processGrouped(ctx, message, from, startLogicalPos)
	}

	// check if finished
	if _, ok := i.finished[startLogicalPos]; ok {
		i.logicalPos++ // increment logical position even if skipped
		return false, true
	}

	ret, skip = i.processSingle(ctx, message, from, startLogicalPos)
	i.logicalPos++ // increment logical position after processing
	return ret, skip
}

func (i *iter) processSingle(ctx context.Context, message *tg.Message, from peers.Peer, logicalPos int) (bool, bool) {
	item, ok := tmedia.GetMedia(message)
	if !ok {
		logctx.From(ctx).Warn("Message has no media",
			zap.Int64("dialog_id", from.ID()),
			zap.Int("message_id", message.ID),
		)

		return false, true
	}

	// process include and exclude
	ext := filepath.Ext(item.Name)
	if _, ok = i.include[ext]; len(i.include) > 0 && !ok {
		return false, true
	}
	if _, ok = i.exclude[ext]; len(i.exclude) > 0 && ok {
		return false, true
	}

	toName := bytes.Buffer{}
	err := i.tpl.Execute(&toName, &fileTemplate{
		DialogID:     from.ID(),
		MessageID:    message.ID,
		MessageDate:  int64(message.Date),
		FileName:     item.Name,
		FileCaption:  message.Message,
		FileSize:     utils.Byte.FormatBinaryBytes(item.Size),
		DownloadDate: time.Now().Unix(),
	})
	if err != nil {
		i.err = errors.Wrap(err, "execute template")
		return false, false
	}

	if i.opts.SkipSame {
		if stat, err := os.Stat(filepath.Join(i.opts.Dir, toName.String())); err == nil {
			if fsutil.GetNameWithoutExt(toName.String()) == fsutil.GetNameWithoutExt(stat.Name()) &&
				stat.Size() == item.Size {
				return false, true
			}
		}
	}

	filename := fmt.Sprintf("%s%s", toName.String(), tempExt)
	path := filepath.Join(i.opts.Dir, filename)

	// #113. If path contains dirs, create it. So now we support nested dirs.
	if err = os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		i.err = errors.Wrap(err, "create dir")
		return false, false
	}

	to, err := os.Create(path)
	if err != nil {
		i.err = errors.Wrap(err, "create file")
		return false, false
	}

	i.elem <- &iterElem{
		id:         int(i.counter.Inc()),
		logicalPos: logicalPos,

		from:    from,
		fromMsg: message,
		file:    item,

		to: to,

		opts: i.opts,
	}

	return true, false
}

func (i *iter) processGrouped(ctx context.Context, message *tg.Message, from peers.Peer, startLogicalPos int) (bool, bool) {
	grouped, err := tutil.GetGroupedMessages(ctx, i.pool.Default(ctx), from.InputPeer(), message)
	if err != nil {
		i.err = errors.Wrapf(err, "resolve grouped message %d/%d", from.ID(), message.ID)
		return false, false
	}

	hasValid := false

	for idx, msg := range grouped {
		logicalPos := startLogicalPos + idx

		// check if this grouped message is already finished
		if _, ok := i.finished[logicalPos]; ok {
			continue
		}

		ret, skip := i.processSingle(ctx, msg, from, logicalPos)

		// if processSingle encounters a fatal error (not just skip), propagate it
		if !ret && !skip {
			// i.err should already be set by processSingle
			return false, false
		}

		if ret {
			hasValid = true
		}
	}

	// increment logical position by the number of messages in the group
	i.logicalPos += len(grouped)

	return hasValid, !hasValid
}

func (i *iter) Value() downloader.Elem {
	return <-i.elem
}

func (i *iter) Err() error {
	return i.err
}

func (i *iter) SetFinished(finished map[int]struct{}) {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.finished = finished
}

func (i *iter) Finished() map[int]struct{} {
	i.mu.Lock()
	defer i.mu.Unlock()

	return i.finished
}

func (i *iter) Fingerprint() string {
	return i.fingerprint
}

func (i *iter) Finish(id int) {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.finished[id] = struct{}{}
}

func (i *iter) Total() int {
	i.mu.Lock()
	defer i.mu.Unlock()

	total := 0
	for _, m := range i.dialogs {
		total += len(m.Messages)
	}
	return total
}

func (i *iter) SkippedDeleted() int64 {
	return i.skippedDeleted.Load()
}

func (i *iter) DeletedIDs() []string {
	i.mu.Lock()
	defer i.mu.Unlock()
	return i.deletedIDs
}

// positionToLogicalIndex converts physical position (dialogIndex, messageIndex) to logical index
// This method is kept for potential future use but is currently unused.
// func (i *iter) positionToLogicalIndex(dialogIdx, messageIdx int) int {
// 	return i.preSum[dialogIdx] + messageIdx
// }

func flatDialogs(dialogs [][]*tmessage.Dialog) []*tmessage.Dialog {
	res := make([]*tmessage.Dialog, 0)
	for _, d := range dialogs {
		if len(d) == 0 {
			continue
		}
		res = append(res, d...)
	}
	return res
}

func sortDialogs(dialogs []*tmessage.Dialog, desc bool) {
	sort.Slice(dialogs, func(i, j int) bool {
		return tutil.GetInputPeerID(dialogs[i].Peer) <
			tutil.GetInputPeerID(dialogs[j].Peer) // increasing order
	})

	for _, m := range dialogs {
		sort.Slice(m.Messages, func(i, j int) bool {
			if desc {
				return m.Messages[i] > m.Messages[j]
			}
			return m.Messages[i] < m.Messages[j]
		})
	}
}

// preSum of dialogs
// This method is kept for potential future use but is currently unused.
// func preSum(dialogs []*tmessage.Dialog) []int {
// 	sum := make([]int, len(dialogs)+1)
// 	for i, m := range dialogs {
// 		sum[i+1] = sum[i] + len(m.Messages)
// 	}
// 	return sum
// }

func fingerprint(dialogs []*tmessage.Dialog) string {
	endian := binary.BigEndian
	buf, b := &bytes.Buffer{}, make([]byte, 8)
	for _, m := range dialogs {
		endian.PutUint64(b, uint64(tutil.GetInputPeerID(m.Peer)))
		buf.Write(b)
		for _, msg := range m.Messages {
			endian.PutUint64(b, uint64(msg))
			buf.Write(b)
		}
	}

	return fmt.Sprintf("%x", sha256.Sum256(buf.Bytes()))
}
