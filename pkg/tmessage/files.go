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
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	peer, err := getChatInfo(ctx, client, kvd, file)
	if err != nil {
		return nil, err
	}
	logctx.From(ctx).Debug("Got peer info",
		zap.Int64("id", peer.ID()),
		zap.String("name", peer.VisibleName()))

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

	// A parse error closes d.Stream() just like the end of input does,
	// so without this check a broken file looks like a short one.
	if err := d.Err(); err != nil {
		return nil, err
	}

	return m, nil
}

func getChatInfo(ctx context.Context, client *tg.Client, kvd storage.Storage, file string) (peers.Peer, error) {
	// Use a file handle of our own. jstream's decoder parses and reads in
	// background goroutines that keep running after we stop receiving from
	// d.Stream(), and it has no way to be stopped. Closing this handle makes
	// its next Read fail, which ends them. On a shared handle they would keep
	// moving the offset that collect() reads from.
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	d := jstream.NewDecoder(f, 1).EmitKV()

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

	// Stop the decoder now, before the network call below.
	_ = f.Close()

	if chatID == 0 {
		return nil, errors.New("can't get chat type or chat id")
	}

	manager := peers.Options{Storage: storage.NewPeers(kvd)}.Build(client)
	return tutil.GetInputPeer(ctx, manager, strconv.FormatInt(chatID, 10))
}
