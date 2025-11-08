package tmessage

import (
	"context"
	"errors"
	"io"
	"os"
	"strconv"

	"github.com/bcicen/jstream"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"

	"github.com/iyear/tdl/core/dcpool"
	"github.com/iyear/tdl/core/logctx"
	"github.com/iyear/tdl/core/storage"
	"github.com/iyear/tdl/core/util/tutil"
)

const (
	keyID       = "id"
	typeMessage = "message"
)

type fMessage struct {
	ID     int            `mapstructure:"id"`
	Type   string         `mapstructure:"type"`
	Time   string         `mapstructure:"date_unixtime"`
	File   string         `mapstructure:"file"`
	Photo  string         `mapstructure:"photo"`
	FromID string         `mapstructure:"from_id"`
	From   string         `mapstructure:"from"`
	Text   interface{}    `mapstructure:"text"`
	Raw    map[string]any `mapstructure:"raw"` // Raw Telegram message data (only present with --raw export)
}

func FromFile(ctx context.Context, pool dcpool.Pool, kvd storage.Storage, files []string, onlyMedia bool) ParseSource {
	return func() ([]*Dialog, error) {
		dialogs := make([]*Dialog, 0, len(files))

		for _, file := range files {
			d, err := parseFile(ctx, pool.Default(ctx), kvd, file, onlyMedia)
			if err != nil {
				return nil, err
			}

			logctx.From(ctx).Debug("Parse file",
				zap.String("file", file),
				zap.Int("num", len(d.Messages)))
			dialogs = append(dialogs, d)
		}

		return dialogs, nil
	}
}

func parseFile(ctx context.Context, client *tg.Client, kvd storage.Storage, file string, onlyMedia bool) (*Dialog, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	peer, err := getChatInfo(ctx, client, kvd, f)
	if err != nil {
		return nil, err
	}
	logctx.From(ctx).Debug("Got peer info",
		zap.Int64("id", peer.ID()),
		zap.String("name", peer.VisibleName()))

	if _, err = f.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	return collect(ctx, f, peer, onlyMedia)
}

func collect(ctx context.Context, r io.Reader, peer peers.Peer, onlyMedia bool) (*Dialog, error) {
	d := jstream.NewDecoder(r, 2)

	m := &Dialog{
		Peer:         peer.InputPeer(),
		Messages:     make([]int, 0),
		MessageMetas: make(map[int]*MessageMeta),
	}

	for mv := range d.Stream() {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			fm := fMessage{}

			if mv.ValueType != jstream.Object {
				continue
			}

			if err := mapstructure.WeakDecode(mv.Value, &fm); err != nil {
				return nil, err
			}

			if fm.ID < 0 || fm.Type != typeMessage {
				continue
			}

			if fm.File == "" && fm.Photo == "" && onlyMedia {
				continue
			}

			m.Messages = append(m.Messages, fm.ID)

			// Store metadata from JSON for skip-same optimization
			// Parse date
			var messageDate int64
			if fm.Time != "" {
				if ts, err := strconv.ParseInt(fm.Time, 10, 64); err == nil {
					messageDate = ts
				}
			}

			// Parse text/caption
			var textContent string
			if fm.Text != nil {
				switch v := fm.Text.(type) {
				case string:
					textContent = v
				case []interface{}:
					// Text can be array of objects with "text" fields
					for _, item := range v {
						if m, ok := item.(map[string]interface{}); ok {
							if t, ok := m["text"].(string); ok {
								textContent += t
							}
						}
					}
				}
			}

			// Store metadata if we have a file or photo
			filename := fm.File
			if filename == "" {
				filename = fm.Photo
			}

			if filename != "" {
				// Check if this message has raw data (from --raw export)
				if fm.Raw != nil && !m.HasRawData {
					m.HasRawData = true
					logctx.From(ctx).Debug("JSON export includes raw Telegram message data")
				}

				m.MessageMetas[fm.ID] = &MessageMeta{
					ID:          fm.ID,
					Filename:    filename,
					Date:        messageDate,
					TextContent: textContent,
				}
				if len(m.MessageMetas) <= 3 {
					logctx.From(ctx).Debug("Stored metadata",
						zap.Int("id", fm.ID),
						zap.String("filename", filename),
						zap.Int64("date", messageDate))
				}
			}
		}
	}

	logctx.From(ctx).Info("Collected messages with metadata",
		zap.Int("total_messages", len(m.Messages)),
		zap.Int("messages_with_metadata", len(m.MessageMetas)),
		zap.Bool("has_raw_data", m.HasRawData))

	return m, nil
}

func getChatInfo(ctx context.Context, client *tg.Client, kvd storage.Storage, r io.Reader) (peers.Peer, error) {
	d := jstream.NewDecoder(r, 1).EmitKV()

	chatID := int64(0)

	for mv := range d.Stream() {
		_kv, ok := mv.Value.(jstream.KV)
		if !ok {
			continue
		}

		if _kv.Key == keyID {
			chatID = int64(_kv.Value.(float64))
		}

		if chatID != 0 {
			break
		}
	}

	if chatID == 0 {
		return nil, errors.New("can't get chat type or chat id")
	}

	manager := peers.Options{Storage: storage.NewPeers(kvd)}.Build(client)
	return tutil.GetInputPeer(ctx, manager, strconv.FormatInt(chatID, 10))
}
