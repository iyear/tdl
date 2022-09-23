package dl

import (
	"context"
	"errors"
	"fmt"
	"github.com/bcicen/jstream"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
	"github.com/mitchellh/mapstructure"
	"io"
	"os"
)

// https://github.com/telegramdesktop/tdesktop/blob/dev/Telegram/SourceFiles/export/output/export_output_json.cpp#L1112-L1124
var typeMap = map[string]uint32{
	"saved_messages":     tg.InputPeerSelfTypeID,
	"personal_chat":      tg.InputPeerUserTypeID,
	"bot_chat":           tg.InputPeerUserTypeID,
	"private_group":      tg.InputPeerChatTypeID,
	"private_supergroup": tg.InputPeerChannelTypeID,
	"public_supergroup":  tg.InputPeerChannelTypeID,
	"private_channel":    tg.InputPeerChannelTypeID,
	"public_channel":     tg.InputPeerChannelTypeID,
}

const (
	keyID       = "id"
	keyType     = "type"
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

func parseFiles(ctx context.Context, client *tg.Client, files []string) ([]*dialog, error) {
	dialogs := make([]*dialog, 0, len(files))

	for _, file := range files {
		d, err := parseFile(ctx, client, file)
		if err != nil {
			return nil, err
		}

		dialogs = append(dialogs, d)
	}

	return dialogs, nil
}

func parseFile(ctx context.Context, client *tg.Client, file string) (*dialog, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	peer, err := getChatInfo(ctx, client, f)
	if err != nil {
		return nil, err
	}

	if _, err = f.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	return collect(ctx, f, peer)
}

func collect(ctx context.Context, r io.Reader, peer tg.InputPeerClass) (*dialog, error) {
	d := jstream.NewDecoder(r, 2)

	m := &dialog{
		peer: peer,
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

func getChatInfo(ctx context.Context, client *tg.Client, r io.Reader) (tg.InputPeerClass, error) {
	d := jstream.NewDecoder(r, 1).EmitKV()

	chatType, chatID := uint32(0), int64(0)

	for mv := range d.Stream() {
		kv, ok := mv.Value.(jstream.KV)
		if !ok {
			continue
		}

		if kv.Key == keyType {
			v := kv.Value.(string)
			chatType, ok = typeMap[v]
			if !ok {
				return nil, fmt.Errorf("unsupported dialog type: %s", v)
			}
		}

		if kv.Key == keyID {
			chatID = int64(kv.Value.(float64))
		}

		if chatType != 0 && chatID != 0 {
			break
		}
	}

	if chatType == 0 || chatID == 0 {
		return nil, errors.New("can't get chat type or chat id")
	}

	var (
		peer peers.Peer
		err  error
	)
	manager := peers.Options{}.Build(client)

	switch chatType {
	case tg.InputPeerSelfTypeID:
		return &tg.InputPeerSelf{}, nil
	case tg.InputPeerUserTypeID:
		peer, err = manager.ResolveUserID(ctx, chatID)
	case tg.InputPeerChatTypeID:
		peer, err = manager.ResolveChatID(ctx, chatID)
	case tg.InputPeerChannelTypeID:
		peer, err = manager.ResolveChannelID(ctx, chatID)
	}

	if err != nil {
		return nil, err
	}

	return peer.InputPeer(), nil
}
