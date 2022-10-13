package dl

import (
	"context"
	"errors"
	"github.com/bcicen/jstream"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
	"github.com/iyear/tdl/pkg/kv"
	"github.com/iyear/tdl/pkg/storage"
	"github.com/iyear/tdl/pkg/utils"
	"github.com/mitchellh/mapstructure"
	"io"
	"os"
	"strconv"
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

func parseFiles(ctx context.Context, client *tg.Client, kvd *kv.KV, files []string) ([]*dialog, error) {
	dialogs := make([]*dialog, 0, len(files))

	for _, file := range files {
		d, err := parseFile(ctx, client, kvd, file)
		if err != nil {
			return nil, err
		}

		dialogs = append(dialogs, d)
	}

	return dialogs, nil
}

func parseFile(ctx context.Context, client *tg.Client, kvd *kv.KV, file string) (*dialog, error) {
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

	if _, err = f.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	return collect(ctx, f, peer)
}

func collect(ctx context.Context, r io.Reader, peer peers.Peer) (*dialog, error) {
	d := jstream.NewDecoder(r, 2)

	m := &dialog{
		peer: peer.InputPeer(),
		msgs: make([]int, 0),
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

			if fm.File == "" && fm.Photo == "" {
				continue
			}

			m.msgs = append(m.msgs, fm.ID)
		}
	}

	return m, nil
}

func getChatInfo(ctx context.Context, client *tg.Client, kvd *kv.KV, r io.Reader) (peers.Peer, error) {
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

	return utils.Telegram.GetInputPeer(ctx, peers.Options{Storage: storage.NewPeers(kvd)}.Build(client), strconv.FormatInt(chatID, 10))
}
