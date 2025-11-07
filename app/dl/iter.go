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
	counter *atomic.Int64
	elem    chan downloader.Elem
	err     error

	// Optimization statistics
	skipSameOptimizationHits *atomic.Int64 // Files skipped using JSON metadata (no network call)
	skipSameNetworkChecks    *atomic.Int64 // Files checked via network calls
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
		logicalPos:   0,
		dialogIndex:  0,
		messageIndex: 0,
		counter:      atomic.NewInt64(-1),
		elem:         make(chan downloader.Elem, 10), // grouped message buffer
		err:          nil,

		skipSameOptimizationHits: atomic.NewInt64(0),
		skipSameNetworkChecks:    atomic.NewInt64(0),
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

	//  Early skip-same optimization: If we have filename metadata from JSON export
	// and skip-same is enabled, check if file exists locally before making network calls
	// Only works with default template pattern
	const defaultTemplate = `{{ .DialogID }}_{{ .MessageID }}_{{ filenamify .FileName }}`
	if i.opts.SkipSame && i.opts.Template == defaultTemplate {
		dialog := i.dialogs[i.dialogIndex]

		// Log optimization availability on first message
		if i.logicalPos == 0 && len(dialog.MessageMetas) > 0 {
			logctx.From(ctx).Info("Skip-same optimization enabled",
				zap.Int("messages_with_metadata", len(dialog.MessageMetas)),
				zap.Int("total_messages", len(dialog.Messages)),
				zap.String("optimization", "JSON metadata available - will skip existing files without network calls"))
		}

		// Debug logging for first few messages
		if i.logicalPos <= 2 {
			logctx.From(ctx).Debug("Early skip-same check",
				zap.Int("logical_pos", i.logicalPos),
				zap.Int("msg_id", msg),
				zap.Int("meta_count", len(dialog.MessageMetas)),
				zap.Bool("has_meta", dialog.MessageMetas[msg] != nil))
		}

		if len(dialog.MessageMetas) > 0 {
			if meta, ok := dialog.MessageMetas[msg]; ok && meta.Filename != "" {
				// We have filename from JSON, construct the expected filename using the template pattern
				// Default template is: {{ .DialogID }}_{{ .MessageID }}_{{ filenamify .FileName }}
				// Extract peer ID from InputPeerClass without network call
				var peerID int64
				switch p := dialog.Peer.(type) {
				case *tg.InputPeerChannel:
					peerID = p.ChannelID
				case *tg.InputPeerUser:
					peerID = p.UserID
				case *tg.InputPeerChat:
					peerID = p.ChatID
				default:
					// Unknown peer type, skip optimization
					if i.logicalPos <= 5 {
						logctx.From(ctx).Warn("Skip-same optimization unavailable for message",
							zap.Int("msg_id", msg),
							zap.String("reason", "unknown peer type"),
							zap.String("peer_type", fmt.Sprintf("%T", dialog.Peer)))
					}
					goto skipOptimization
				}

				// Construct expected filename: {DialogID}_{MessageID}_{FileName}
				expectedFilename := fmt.Sprintf("%d_%d_%s", peerID, msg, meta.Filename)
				checkPath := filepath.Join(i.opts.Dir, expectedFilename)

				if stat, err := os.Stat(checkPath); err == nil {
					// File exists, check if name (without ext) matches
					if fsutil.GetNameWithoutExt(expectedFilename) == fsutil.GetNameWithoutExt(stat.Name()) {
						// File with same name exists, skip without network call
						i.skipSameOptimizationHits.Inc()
						if i.logicalPos <= 5 || i.skipSameOptimizationHits.Load()%100 == 0 {
							logctx.From(ctx).Info("Skipped existing file (no network call)",
								zap.String("file", expectedFilename),
								zap.Int64("size_bytes", stat.Size()),
								zap.Int64("optimization_hits", i.skipSameOptimizationHits.Load()))
						}
						i.logicalPos++
						return false, true
					} else {
						// Name mismatch, fall through to network check
						if i.logicalPos <= 5 {
							logctx.From(ctx).Debug("File exists but name mismatch",
								zap.String("expected", expectedFilename),
								zap.String("found", stat.Name()))
						}
					}
				} else if !os.IsNotExist(err) {
					// Stat error (not just file not found), log it
					logctx.From(ctx).Warn("Error checking file existence",
						zap.String("file", checkPath),
						zap.Error(err))
				}
				// File doesn't exist, will proceed to network check below
				i.skipSameNetworkChecks.Inc()
			}
		} else if i.logicalPos == 0 {
			logctx.From(ctx).Warn("Skip-same optimization unavailable",
				zap.String("reason", "no metadata in JSON export"),
				zap.String("note", "all files will require network calls to check"))
		}
	} else if i.opts.SkipSame && i.opts.Template != defaultTemplate && i.logicalPos == 0 {
		logctx.From(ctx).Warn("Skip-same optimization disabled",
			zap.String("reason", "custom name template in use"),
			zap.String("note", "optimization only works with default template"))
	}
skipOptimization:

	from, err := i.manager.FromInputPeer(ctx, peer)
	if err != nil {
		i.err = errors.Wrap(err, "resolve from input peer")
		return false, false
	}
	message, err := tutil.GetSingleMessage(ctx, i.pool.Default(ctx), peer, msg)
	if err != nil {
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

	ret, skip = i.processSingle(message, from, startLogicalPos)
	i.logicalPos++ // increment logical position after processing
	return ret, skip
}

func (i *iter) processSingle(message *tg.Message, from peers.Peer, logicalPos int) (bool, bool) {
	item, ok := tmedia.GetMedia(message)
	if !ok {
		i.err = errors.Errorf("can not get media from %d/%d message", from.ID(), message.ID)
		return false, false
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

		ret, skip := i.processSingle(msg, from, logicalPos)

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

func (i *iter) GetOptimizationStats() (hits, networkChecks int64) {
	if i.skipSameOptimizationHits == nil || i.skipSameNetworkChecks == nil {
		return 0, 0
	}
	return i.skipSameOptimizationHits.Load(), i.skipSameNetworkChecks.Load()
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
