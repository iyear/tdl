package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"math"
	"os"
	"time"

	"github.com/expr-lang/expr"
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

loop:
	for msg, err := range exportHistory(ctx, q, opts) {
		if err != nil {
			return err
		}
		switch opts.Type {
		case ExportTypeTime:
			if msg.GetDate() < opts.Input[0] {
				break loop
			}
			if msg.GetDate() > opts.Input[1] {
				continue
			}
		case ExportTypeId:
			if msg.GetID() < opts.Input[0] {
				break loop
			}
		case ExportTypeLast:
			if count >= int64(opts.Input[0]) {
				break loop
			}
		}

		m, ok := msg.(*tg.Message)
		if !ok {
			continue
		}
		// only get media messages
		media, ok := tmedia.GetMedia(m)
		if !ok && !opts.All {
			continue
		}

		b, err := texpr.Run(filter, texpr.ConvertEnvMessage(m))
		if err != nil {
			return fmt.Errorf("failed to run filter: %w", err)
		}
		if !b.(bool) { // filtered
			continue
		}

		fileName := ""
		if media != nil { // #207
			fileName = media.Name
		}
		t := &Message{
			ID:   m.ID,
			Type: "message",
			File: fileName,
		}
		if opts.WithContent {
			t.Date = m.Date
			t.Text = m.Message
		}
		if opts.Raw {
			t.Raw = m
		}

		mb, err := json.Marshal(t)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %w", err)
		}
		enc.Raw(mb)

		count++
		tracker.SetValue(count)
	}

	tracker.MarkAsDone()
	prog.Wait(ctx, pw)
	return nil
}

func exportHistory(ctx context.Context, q messages.Query, opts ExportOptions) iter.Seq2[tg.NotEmptyMessage, error] {
	req := messages.Request{Limit: 100}
	switch opts.Type {
	case ExportTypeTime:
		if end := opts.Input[1]; end < math.MaxInt32 {
			req.OffsetDate = end + 1 // Telegram's upper bound is exclusive.
		}
	case ExportTypeId:
		if end := opts.Input[1]; end < math.MaxInt32 {
			req.OffsetID = end + 1 // #89: retain the last msg id
		}
	case ExportTypeLast:
		// Start at the newest message without an offset.
	}

	return func(yield func(tg.NotEmptyMessage, error) bool) {
		cursor := req
		for {
			batch, err := fetchHistoryPage(ctx, q, cursor)
			if err != nil {
				yield(nil, err)
				return
			}
			// A short batch can still have more history behind it.
			if len(batch) == 0 {
				return
			}
			nextID := batch[len(batch)-1].GetID()
			if nextID <= 0 || (cursor.OffsetID != 0 && nextID >= cursor.OffsetID) {
				yield(nil, fmt.Errorf("message pagination did not advance: offset %d, next offset %d", cursor.OffsetID, nextID))
				return
			}
			for _, item := range batch {
				if msg, ok := item.AsNotEmpty(); ok && !yield(msg, nil) {
					return
				}
			}
			// Use the date only to seek, then page by ID so messages with
			// the same timestamp as the cursor aren't excluded.
			cursor.OffsetDate = 0
			cursor.OffsetID = nextID
		}
	}
}

func fetchHistoryPage(ctx context.Context, q messages.Query, req messages.Request) (tg.MessageClassArray, error) {
	r, err := q.Query(ctx, req)
	if err != nil {
		return nil, err
	}

	var raw tg.MessageClassArray
	switch result := r.(type) {
	case *tg.MessagesMessages:
		raw = result.Messages
	case *tg.MessagesMessagesSlice:
		raw = result.Messages
	case *tg.MessagesChannelMessages:
		raw = result.Messages
	default:
		return nil, fmt.Errorf("unexpected message response type %T", r)
	}

	return raw.SortStable(func(a, b tg.MessageClass) bool {
		return a.GetID() > b.GetID()
	}), nil
}
