package dl

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/tg"
	"github.com/iyear/tdl/pkg/downloader"
	"github.com/iyear/tdl/pkg/kv"
	"github.com/iyear/tdl/pkg/storage"
	"github.com/iyear/tdl/pkg/utils"
	"path/filepath"
	"sync"
	"text/template"
	"time"
)

type iter struct {
	client           *tg.Client
	dialogs          []*dialog
	include, exclude map[string]struct{}
	mu               sync.Mutex
	curi             int
	curj             int
	template         *template.Template
	manager          *peers.Manager
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

func newIter(client *tg.Client, kvd kv.KV, tmpl string, include, exclude []string, items ...[]*dialog) (*iter, error) {
	t, err := template.New("dl").Parse(tmpl)
	if err != nil {
		return nil, err
	}

	mm := make([]*dialog, 0)

	for _, m := range items {
		if len(m) == 0 {
			continue
		}
		mm = append(mm, m...)
	}

	// if msgs is empty, return error to avoid range out of index
	if len(mm) == 0 {
		return nil, fmt.Errorf("you must specify at least one message")
	}

	// include and exclude
	includeMap := make(map[string]struct{})
	for _, v := range include {
		includeMap[utils.FS.AddPrefixDot(v)] = struct{}{}
	}
	excludeMap := make(map[string]struct{})
	for _, v := range exclude {
		excludeMap[utils.FS.AddPrefixDot(v)] = struct{}{}
	}

	return &iter{
		client:   client,
		dialogs:  mm,
		include:  includeMap,
		exclude:  excludeMap,
		curi:     0,
		curj:     -1,
		template: t,
		manager:  peers.Options{Storage: storage.NewPeers(kvd)}.Build(client),
	}, nil
}

func (i *iter) Next(ctx context.Context) (*downloader.Item, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	i.mu.Lock()
	i.curj++
	if i.curj >= len(i.dialogs[i.curi].msgs) {
		if i.curi++; i.curi >= len(i.dialogs) {
			return nil, errors.New("no more items")
		}
		i.curj = 0
	}

	curi := i.dialogs[i.curi]
	cur := curi.msgs[i.curj]
	i.mu.Unlock()

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

	// process include and exclude
	ext := filepath.Ext(media.Name)
	if len(i.include) > 0 {
		if _, ok = i.include[ext]; !ok {
			return nil, downloader.ErrSkip
		}
	}
	if len(i.exclude) > 0 {
		if _, ok = i.exclude[ext]; ok {
			return nil, downloader.ErrSkip
		}
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
	i.mu.Lock()
	defer i.mu.Unlock()

	total := 0
	for _, m := range i.dialogs {
		total += len(m.msgs)
	}
	return total
}
