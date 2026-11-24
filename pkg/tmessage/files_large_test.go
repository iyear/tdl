package tmessage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"

	"github.com/iyear/tdl/core/storage"
)

// TestParseFileLargeExport feeds parseFile (the code behind `tdl dl -f`) a file
// shaped like Telegram Desktop's result.json with 2000 photo messages and checks
// that all of them are collected. No network: the Telegram client is a stub
// that only answers the chat lookup getChatInfo needs.
func TestParseFileLargeExport(t *testing.T) {
	const chatID, n = 123, 2000

	path := filepath.Join(t.TempDir(), "result.json")
	writeDesktopExport(t, path, chatID, n)

	client := tg.NewClient(stubInvoker{chatID: chatID})

	for run := 1; run <= 5; run++ {
		d, err := parseFile(context.Background(), client, memKV{}, path, true)
		if err != nil {
			t.Fatalf("run %d: parseFile: %v", run, err)
		}
		if len(d.Messages) != n {
			t.Errorf("run %d: collected %d of %d messages", run, len(d.Messages), n)
		}
	}
}

// stubInvoker answers messages.getChats with a basic group; every other call fails,
// which is what makes tutil.GetInputPeer fall through to ResolveChatID.
type stubInvoker struct{ chatID int64 }

func (s stubInvoker) Invoke(_ context.Context, input bin.Encoder, output bin.Decoder) error {
	if _, ok := input.(*tg.MessagesGetChatsRequest); !ok {
		return tgerr.New(400, "PEER_ID_INVALID")
	}
	res := &tg.MessagesChats{Chats: []tg.ChatClass{&tg.Chat{
		ID:    s.chatID,
		Title: "Test chat",
		Photo: &tg.ChatPhotoEmpty{},
	}}}
	var b bin.Buffer
	if err := res.Encode(&b); err != nil {
		return err
	}
	return output.Decode(&b)
}

type memKV map[string][]byte

func (m memKV) Get(_ context.Context, key string) ([]byte, error) {
	if v, ok := m[key]; ok {
		return v, nil
	}
	return nil, storage.ErrNotFound
}
func (m memKV) Set(_ context.Context, key string, value []byte) error { m[key] = value; return nil }
func (m memKV) Delete(_ context.Context, key string) error            { delete(m, key); return nil }

func writeDesktopExport(t *testing.T, path string, chatID int64, n int) {
	t.Helper()
	type msg struct {
		ID           int    `json:"id"`
		Type         string `json:"type"`
		Date         string `json:"date"`
		DateUnixtime string `json:"date_unixtime"`
		From         string `json:"from"`
		FromID       string `json:"from_id"`
		Photo        string `json:"photo"`
		Width        int    `json:"width"`
		Height       int    `json:"height"`
		Text         string `json:"text"`
	}
	msgs := make([]msg, n)
	for i := range msgs {
		msgs[i] = msg{
			ID: i + 1, Type: "message", Date: "2021-01-01T12:00:00", DateUnixtime: "1609491600",
			From: "Alice", FromID: "user1", Photo: fmt.Sprintf("photos/photo_%d.jpg", i+1),
			Width: 1280, Height: 960,
		}
	}
	out := struct {
		Name     string `json:"name"`
		Type     string `json:"type"`
		ID       int64  `json:"id"`
		Messages []msg  `json:"messages"`
	}{"Test chat", "private_group", chatID, msgs}
	b, err := json.MarshalIndent(out, "", " ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatal(err)
	}
}
