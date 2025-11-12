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
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/fatih/color"
	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
	pw "github.com/jedib0t/go-pretty/v6/progress"
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
	pw      pw.Writer

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
	opts Options, delay time.Duration, pw pw.Writer,
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

	// Check if all dialogs have zero messages
	totalMessages := 0
	for _, d := range dialogs {
		totalMessages += len(d.Messages)
	}
	if totalMessages == 0 {
		return nil, errors.Errorf("no messages found in provided source (all dialogs contain 0 messages)")
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
		pw:      pw,

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

	// Quick skip-same optimization: Uses JSON metadata to check files without network calls.
	// Enabled when:
	// 1. --skip-same flag is set
	// 2. --force-web-check is NOT set
	// 3. Either:
	//    a) JSON export includes raw Telegram data (--raw flag, works with any template), OR
	//    b) Using default template with standard JSON export
	if i.opts.SkipSame && !i.opts.ForceWebCheck {
		if skipped := i.trySkipSameOptimization(ctx, msg); skipped {
			i.logicalPos++
			return false, true
		}
	}

	from, err := i.manager.FromInputPeer(ctx, peer)
	if err != nil {
		i.err = errors.Wrap(err, "resolve from input peer")
		return false, false
	}
	message, err := tutil.GetSingleMessage(ctx, i.pool.Default(ctx), peer, msg)
	if err != nil {
		// Check if message was deleted using proper error type
		var deletedErr *tutil.DeletedMessageError
		if errors.As(err, &deletedErr) {
			color.Red("Message no longer exists: %d/%d", deletedErr.PeerID, deletedErr.MessageID)
			i.logicalPos++ // increment logical position to skip this message
			return false, true
		}
		// Check if message is an unsupported type (MessageService, MessageEmpty, etc.)
		var unsupportedErr *tutil.UnsupportedMessageTypeError
		if errors.As(err, &unsupportedErr) {
			color.Yellow("Skipping system message: %d/%d (%s)", 
				unsupportedErr.PeerID, unsupportedErr.MessageID, unsupportedErr.MessageType)
			i.logicalPos++ // increment logical position to skip this message
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
	item, isSupportedMediaType, err := tmedia.GetMedia(ctx, message)
	if err != nil {
		i.err = errors.Wrap(err, fmt.Sprintf("can not get media from %d/%d message", from.ID(), message.ID))
		return false, false
	}
	if !isSupportedMediaType {
		logctx.
			From(ctx).
			Info("unsupported media type",
				zap.Int64("dialog", from.ID()),
				zap.Int("message", message.ID),
				zap.String("media_type", fmt.Sprintf("%T", message.Media)))

		i.pw.Log(color.YellowString("Skip unsupported media type %T in %s - peer id: %d - message id: %d",
			message.Media, from.VisibleName(), from.ID(), message.ID))
		return false, true
	}

	// process include and exclude
	ext := filepath.Ext(item.Name)
	if _, ok := i.include[ext]; len(i.include) > 0 && !ok {
		return false, true
	}
	if _, ok := i.exclude[ext]; len(i.exclude) > 0 && ok {
		return false, true
	}

	toName := bytes.Buffer{}
	err = i.tpl.Execute(&toName, &fileTemplate{
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

// trySkipSameOptimization attempts to skip downloading a file using JSON metadata
// without making network calls. Returns true if the file was skipped.
func (i *iter) trySkipSameOptimization(ctx context.Context, msg int) bool {
	dialog := i.dialogs[i.dialogIndex]
	const defaultTemplate = `{{ .DialogID }}_{{ .MessageID }}_{{ filenamify .FileName }}`
	isDefaultTemplate := i.opts.Template == defaultTemplate
	canOptimize := dialog.HasRawData || isDefaultTemplate

	// Log optimization status on first message
	if i.logicalPos == 0 {
		// Warn if using custom template without MessageID (collision risk)
		if canOptimize && !strings.Contains(i.opts.Template, "MessageID") {
			logctx.From(ctx).Warn("Template does not include MessageID - filename collisions may occur",
				zap.String("current_template", i.opts.Template),
				zap.String("recommendation", "Include {{ .MessageID }} in template to ensure unique filenames"),
				zap.String("note", "Files with duplicate names will be skipped by --skip-same"))
		}

		if !canOptimize {
			logctx.From(ctx).Warn("Skip-same optimization disabled",
				zap.String("reason", "requires either raw JSON export or default template"),
				zap.String("solution", "Use --raw flag during export, or use default template"),
				zap.Bool("has_raw_data", dialog.HasRawData),
				zap.Bool("is_default_template", isDefaultTemplate),
				zap.String("note", "Falling back to network-based file checking"))
		} else if len(dialog.MessageMetas) > 0 {
			logctx.From(ctx).Info("Skip-same optimization enabled",
				zap.Int("messages_with_metadata", len(dialog.MessageMetas)),
				zap.Int("total_messages", len(dialog.Messages)),
				zap.Bool("has_raw_data", dialog.HasRawData),
				zap.Bool("is_default_template", isDefaultTemplate),
				zap.String("note", "Using JSON metadata to skip files without network calls. Use --force-web-check to disable."))
		} else {
			logctx.From(ctx).Warn("Skip-same optimization unavailable",
				zap.String("reason", "no metadata in JSON export"),
				zap.String("note", "Files will require network calls to check"))
		}
	}

	// Only proceed with optimization if requirements are met
	if !canOptimize || len(dialog.MessageMetas) == 0 {
		return false
	}

	meta, ok := dialog.MessageMetas[msg]
	if !ok || meta.Filename == "" {
		return false
	}

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
		// Unknown peer type, skip optimization for this message
		if i.logicalPos <= 3 {
			logctx.From(ctx).Debug("Quick skip-same: unknown peer type, using network check",
				zap.String("peer_type", fmt.Sprintf("%T", dialog.Peer)))
		}
		i.skipSameNetworkChecks.Inc()
		return false
	}

	// Execute template with metadata to construct expected filename
	var expectedFilename strings.Builder
	templateData := &fileTemplate{
		DialogID:    peerID,
		MessageID:   meta.ID,
		MessageDate: meta.Date,
		FileName:    meta.Filename,
		FileCaption: meta.TextContent,
		// FileSize and DownloadDate not available from metadata alone
	}

	if err := i.tpl.Execute(&expectedFilename, templateData); err != nil {
		// Template execution failed, log and fall through to network check
		if i.logicalPos <= 3 {
			logctx.From(ctx).Warn("Quick skip-same: template execution failed",
				zap.Error(err))
		}
		i.skipSameNetworkChecks.Inc()
		return false
	}

	checkPath := filepath.Join(i.opts.Dir, expectedFilename.String())

	stat, err := os.Stat(checkPath)
	if err != nil {
		if !os.IsNotExist(err) {
			// Stat error (not just file not found), log it
			logctx.From(ctx).Warn("Error checking file existence",
				zap.String("file", checkPath),
				zap.Error(err))
		}
		// File doesn't exist, will proceed to network check
		i.skipSameNetworkChecks.Inc()
		return false
	}

	// File exists, check if name (without ext) matches
	if fsutil.GetNameWithoutExt(expectedFilename.String()) == fsutil.GetNameWithoutExt(stat.Name()) {
		// File with same name exists, skip without network call
		i.skipSameOptimizationHits.Inc()
		if i.logicalPos <= 3 || i.skipSameOptimizationHits.Load()%100 == 0 {
			logctx.From(ctx).Info("Skipped existing file (no network call)",
				zap.String("file", expectedFilename.String()),
				zap.Int64("size_bytes", stat.Size()),
				zap.Int64("total_skipped", i.skipSameOptimizationHits.Load()))
		}
		return true
	}

	// Name mismatch, fall through to network check
	if i.logicalPos <= 3 {
		logctx.From(ctx).Debug("File exists but name mismatch",
			zap.String("expected", expectedFilename.String()),
			zap.String("found", stat.Name()))
	}
	i.skipSameNetworkChecks.Inc()
	return false
}
