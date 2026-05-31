package tmessage

import (
	"context"
	"encoding/json"
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
	ID     int         `mapstructure:"id"`
	Type   string      `mapstructure:"type"`
	Time   string      `mapstructure:"date_unixtime"`
	File   string      `mapstructure:"file"`
	Photo  string      `mapstructure:"photo"`
	FromID string      `mapstructure:"from_id"`
	From   string      `mapstructure:"from"`
	Text   interface{} `mapstructure:"text"`
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
	peer, err := getChatInfo(ctx, client, kvd, file)
	if err != nil {
		return nil, err
	}
	logctx.From(ctx).Debug("Got peer info",
		zap.Int64("id", peer.ID()),
		zap.String("name", peer.VisibleName()))

	return collectFile(ctx, file, peer, onlyMedia)
}

func collectFile(ctx context.Context, file string, peer peers.Peer, onlyMedia bool) (*Dialog, error) {
	// Use a fresh handle so chat ID probing cannot race with jstream's streaming decoder on file offsets.
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	return collect(ctx, f, peer, onlyMedia)
}

func collect(ctx context.Context, r io.Reader, peer peers.Peer, onlyMedia bool) (*Dialog, error) {
	d := jstream.NewDecoder(r, 2)

	m := &Dialog{
		Peer:     peer.InputPeer(),
		Messages: make([]int, 0),
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
		}
	}

	return m, nil
}

func getChatInfo(ctx context.Context, client *tg.Client, kvd storage.Storage, file string) (peers.Peer, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	chatID, err := readChatID(f)
	if err != nil {
		return nil, err
	}
	if chatID == 0 {
		return nil, errors.New("can't get chat type or chat id")
	}

	manager := peers.Options{Storage: storage.NewPeers(kvd)}.Build(client)
	return tutil.GetInputPeer(ctx, manager, strconv.FormatInt(chatID, 10))
}

func readChatID(r io.Reader) (int64, error) {
	d := json.NewDecoder(r)
	d.UseNumber()

	tok, err := d.Token()
	if err != nil {
		return 0, err
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '{' {
		return 0, errors.New("expected telegram export JSON object")
	}

	for d.More() {
		tok, err = d.Token()
		if err != nil {
			return 0, err
		}

		k, ok := tok.(string)
		if !ok {
			return 0, errors.New("expected telegram export JSON object key")
		}

		if k != keyID {
			if err = skipJSONValue(d); err != nil {
				return 0, err
			}
			continue
		}

		return decodeChatID(d)
	}

	return 0, errors.New("can't get chat type or chat id")
}

func decodeChatID(d *json.Decoder) (int64, error) {
	tok, err := d.Token()
	if err != nil {
		return 0, err
	}

	switch v := tok.(type) {
	case json.Number:
		return v.Int64()
	case string:
		return strconv.ParseInt(v, 10, 64)
	default:
		return 0, errors.New("invalid telegram export chat id")
	}
}

func skipJSONValue(d *json.Decoder) error {
	tok, err := d.Token()
	if err != nil {
		return err
	}

	delim, ok := tok.(json.Delim)
	if !ok {
		return nil
	}

	switch delim {
	case '{':
		for d.More() {
			if _, err = d.Token(); err != nil {
				return err
			}
			if err = skipJSONValue(d); err != nil {
				return err
			}
		}
		end, err := d.Token()
		if err != nil {
			return err
		}
		if end != json.Delim('}') {
			return errors.New("invalid JSON object")
		}
	case '[':
		for d.More() {
			if err = skipJSONValue(d); err != nil {
				return err
			}
		}
		end, err := d.Token()
		if err != nil {
			return err
		}
		if end != json.Delim(']') {
			return errors.New("invalid JSON array")
		}
	default:
		return errors.New("unexpected JSON delimiter")
	}

	return nil
}
