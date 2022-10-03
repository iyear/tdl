package dl

import (
	"bytes"
	"context"
	"fmt"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/tg"
	"github.com/iyear/tdl/pkg/downloader"
	"github.com/iyear/tdl/pkg/utils"
	"text/template"
	"time"
)

type iter struct {
	client   *tg.Client
	dialogs  []*dialog
	curi     int
	curj     int
	template *template.Template
	manager  *peers.Manager
}

type dialog struct {
	peer tg.InputPeerClass
	msgs []int
}

type fileTemplate struct {
	DialogID     int64
	MessageID    int
	MessageDate  int64
	FileName     string
	FileSize     string
	DownloadDate int64
}

func newIter(client *tg.Client, tmpl string, items ...[]*dialog) (*iter, error) {
	t, err := template.New("dl").Parse(tmpl)
	if err != nil {
		return nil, err
	}

	mm := make([]*dialog, 0)

	for _, m := range items {
		mm = append(mm, m...)
	}

	return &iter{
		client:   client,
		dialogs:  mm,
		curi:     0,
		curj:     -1,
		template: t,
		manager:  peers.Options{}.Build(client), // TODO(iyear): use local cache to speed up
	}, nil
}

func (i *iter) Next(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	default:
	}

	i.curj++
	if i.curj >= len(i.dialogs[i.curi].msgs) {
		if i.curi++; i.curi >= len(i.dialogs) {
			return false
		}
		i.curj = 0
		return true
	}

	return true
}

func (i *iter) Value(ctx context.Context) (*downloader.Item, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	curi := i.dialogs[i.curi]
	cur := curi.msgs[i.curj]

	return i.item(ctx, curi.peer, cur)
}

func (i *iter) item(ctx context.Context, peer tg.InputPeerClass, msg int) (*downloader.Item, error) {
	it := query.Messages(i.client).GetHistory(peer).OffsetID(msg + 1).BatchSize(1).Iter()
	id := utils.Telegram.GetInputPeerID(peer)

	// get one message
	if !it.Next(ctx) {
		return nil, it.Err()
	}

	message, ok := it.Value().Msg.(*tg.Message)
	if !ok {
		return nil, fmt.Errorf("msg is not *tg.Message")
	}

	// check again to avoid deleted message
	if message.ID != msg {
		return nil, fmt.Errorf("msg may be deleted, id: %d", msg)
	}

	media, ok := GetMedia(message)
	if !ok {
		return nil, fmt.Errorf("can not get media info: %d/%d",
			id, message.ID)
	}

	buf := bytes.Buffer{}
	err := i.template.Execute(&buf, &fileTemplate{
		DialogID:     id,
		MessageID:    message.ID,
		MessageDate:  int64(message.Date),
		FileName:     media.Name,
		FileSize:     utils.Byte.FormatBinaryBytes(media.Size),
		DownloadDate: time.Now().Unix(),
	})
	if err != nil {
		return nil, err
	}
	media.Name = buf.String()

	return media, nil
}

func (i *iter) Total(_ context.Context) int {
	return len(i.dialogs)
}
