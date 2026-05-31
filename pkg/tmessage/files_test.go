package tmessage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gotd/td/constant"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
	"github.com/stretchr/testify/require"
)

func TestReadChatIDDoesNotInterfereWithCollect(t *testing.T) {
	ctx := context.Background()
	path := writeTelegramExport(t, 5000)

	f, err := os.Open(path)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, f.Close())
	}()

	id, err := readChatID(f)
	require.NoError(t, err)
	require.Equal(t, int64(123456), id)

	_, err = f.Seek(0, 0)
	require.NoError(t, err)

	dialog, err := collect(ctx, f, testPeer{id: id}, true)
	require.NoError(t, err)
	require.Len(t, dialog.Messages, 5000)
	require.Equal(t, 1, dialog.Messages[0])
	require.Equal(t, 5000, dialog.Messages[4999])
}

func TestCollectFileIsStableAcrossRepeatedReads(t *testing.T) {
	ctx := context.Background()
	path := writeTelegramExport(t, 5000)
	peer := testPeer{id: 123456}

	for range 10 {
		dialog, err := collectFile(ctx, path, peer, true)
		require.NoError(t, err)
		require.Len(t, dialog.Messages, 5000)
	}
}

func TestReadChatIDSkipsTopLevelValues(t *testing.T) {
	id, err := readChatID(strings.NewReader(`{"messages":[{"id":1,"type":"message","file":"1.jpg"}],"name":"test","id":"789"}`))
	require.NoError(t, err)
	require.Equal(t, int64(789), id)
}

func writeTelegramExport(t *testing.T, messages int) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "export.json")
	f, err := os.Create(path)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, f.Close())
	}()

	_, err = fmt.Fprintf(f, `{"name":"test","type":"private_group","id":123456,"messages":[`)
	require.NoError(t, err)

	for i := 1; i <= messages; i++ {
		if i > 1 {
			_, err = fmt.Fprint(f, ",")
			require.NoError(t, err)
		}
		_, err = fmt.Fprintf(f, `{"id":%d,"type":"message","date_unixtime":"1710000000","file":"photos/%d.jpg","text":""}`, i, i)
		require.NoError(t, err)
	}

	_, err = fmt.Fprint(f, `]}`)
	require.NoError(t, err)

	return path
}

type testPeer struct {
	id int64
}

func (p testPeer) ID() int64 {
	return p.id
}

func (p testPeer) TDLibPeerID() constant.TDLibPeerID {
	return constant.TDLibPeerID(p.id)
}

func (p testPeer) VisibleName() string {
	return "test"
}

func (p testPeer) Username() (string, bool) {
	return "", false
}

func (p testPeer) Restricted() ([]tg.RestrictionReason, bool) {
	return nil, false
}

func (p testPeer) Verified() bool {
	return false
}

func (p testPeer) Scam() bool {
	return false
}

func (p testPeer) Fake() bool {
	return false
}

func (p testPeer) InputPeer() tg.InputPeerClass {
	return &tg.InputPeerChat{ChatID: p.id}
}

func (p testPeer) Sync(context.Context) error {
	return nil
}

func (p testPeer) Manager() *peers.Manager {
	return nil
}

func (p testPeer) Report(context.Context, tg.ReportReasonClass, string) error {
	return nil
}

func (p testPeer) Photo(context.Context) (*tg.Photo, bool, error) {
	return nil, false, nil
}
