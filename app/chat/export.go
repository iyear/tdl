package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/expr-lang/expr"
	"github.com/fatih/color"
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
	Update      bool
}

type Message struct {
	ID      int         `json:"id"`
	Type    string      `json:"type"`
	File    string      `json:"file"`
	Date    int         `json:"date,omitempty"`
	Text    string      `json:"text,omitempty"`
	AlbumID int64       `json:"album_id,omitempty"`
	Raw     *tg.Message `json:"raw,omitempty"`
}

// ExportType
// ENUM(time, id, last)
type ExportType int

type ExportFile struct {
	ID       int64     `json:"id"`
	Messages []Message `json:"messages"`
}

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

	var existingMessages []Message
	if opts.Update {
		if _, err := os.Stat(opts.Output); err == nil {
			data, err := os.ReadFile(opts.Output)
			if err != nil {
				return fmt.Errorf("failed to read existing export file: %w", err)
			}

			if len(data) > 0 {
				var existingFile ExportFile
				if err := json.Unmarshal(data, &existingFile); err != nil {
					return fmt.Errorf("failed to parse existing export file: %w", err)
				}

				if existingFile.ID != peer.ID() {
					return fmt.Errorf("chat ID mismatch: existing file has ID %d, current chat has ID %d", existingFile.ID, peer.ID())
				}

				// Find the latest timestamp to determine where to start adding new messages
				if len(existingFile.Messages) > 0 {
					var latestTimestamp int

					for _, msg := range existingFile.Messages {
						if msg.Date > latestTimestamp {
							latestTimestamp = msg.Date
						}
					}

					if latestTimestamp > 0 {
						if opts.Type == ExportTypeTime {
							opts.Input[0] = latestTimestamp + 1
							color.Green("Adding messages from timestamp %d", opts.Input[0])
						}
					}
				}
				existingMessages = existingFile.Messages
			} else {
				color.Yellow("Output file is empty, creating a new one.")
			}
		} else {
			color.Yellow("Output file doesn't exist, creating a new one")
		}
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

	// Collect all messages and organize by album
	var allMessages []Message
	albumMessages := make(map[int64][]Message)
	rawMessages := make(map[int]*tg.Message)
	count := int64(0)

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
		// only get media messages
		media, ok := tmedia.GetMedia(m)
		if !ok && !opts.All {
			continue
		}

		fileName := ""
		if media != nil { // #207
			fileName = media.Name
		}
		t := Message{
			ID:   m.ID,
			Type: "message",
			File: fileName,
		}
		if opts.WithContent {
			t.Date = m.Date
			t.Text = m.Message

			if AlbumID, ok := m.GetGroupedID(); ok {
				t.AlbumID = AlbumID
				albumMessages[AlbumID] = append(albumMessages[AlbumID], t)
			}
		}
		if opts.Raw {
			t.Raw = m
		}

		allMessages = append(allMessages, t)
		rawMessages[m.ID] = m
		count++
		tracker.SetValue(count)
	}

	matchedAlbums := make(map[int64]bool)
	matchedSingleMessages := []Message{}
	processedAlbumIDs := make(map[int64]bool)

	for _, msg := range allMessages {
		rawMsg, exists := rawMessages[msg.ID]
		if !exists {
			continue
		}

		// Apply filter
		b, err := texpr.Run(filter, texpr.ConvertEnvMessage(rawMsg))
		if err != nil {
			return fmt.Errorf("failed to run filter: %w", err)
		}

		if b.(bool) { // Message matches filter
			if msg.AlbumID != 0 {
				matchedAlbums[msg.AlbumID] = true
			} else {
				matchedSingleMessages = append(matchedSingleMessages, msg)
			}
		}
	}

	// Construct the final list of messages
	var newMessages []Message
	processedIDs := make(map[int]bool) // Track processed message IDs to avoid duplicates

	// Add matched single messages first
	for _, msg := range matchedSingleMessages {
		if !processedIDs[msg.ID] {
			newMessages = append(newMessages, msg)
			processedIDs[msg.ID] = true
		}
	}

	// Add all messages from matched albums
	includedAlbumMessagesCount := 0
	for albumID := range matchedAlbums {
		// if processedAlbumIDs[albumID] { // Skip if album is already processed
		// 	continue
		// }
		albumMessagesAdded := 0
		for _, msg := range albumMessages[albumID] {
			if !processedIDs[msg.ID] {
				newMessages = append(newMessages, msg)
				processedIDs[msg.ID] = true
				albumMessagesAdded++
			}
		}
		if albumMessagesAdded > 0 {
			includedAlbumMessagesCount += albumMessagesAdded
			processedAlbumIDs[albumID] = true
		}
	}

	if err = iter.Err(); err != nil {
		return err
	}

	tracker.SetValue(int64(len(newMessages)))
	tracker.MarkAsDone()
	prog.Wait(ctx, pw)

	var finalMessages []Message
	if opts.Update && len(existingMessages) > 0 {
		finalMessages = append(newMessages, existingMessages...)
		color.Green("Added %d new messages to %d existing messages", len(newMessages), len(existingMessages))
	} else {
		finalMessages = newMessages
	}

	if !opts.WithContent {
		for i := range finalMessages {
			finalMessages[i].AlbumID = 0
		}
	}

	exportData := ExportFile{
		ID:       id,
		Messages: finalMessages,
	}

	jsonData, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal final JSON data: %w", err)
	}

	// Create output directory if it doesn't exist
	dir := filepath.Dir(opts.Output)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	f, err := os.Create(opts.Output)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer multierr.AppendInvoke(&rerr, multierr.Close(f))

	if _, err := f.Write(jsonData); err != nil {
		return fmt.Errorf("failed to write JSON data to file: %w", err)
	}

	return nil
}
