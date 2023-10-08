package dliter

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"text/template"
	"time"

	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/tg"

	"github.com/iyear/tdl/pkg/downloader"
	"github.com/iyear/tdl/pkg/storage"
	"github.com/iyear/tdl/pkg/tmedia"
	"github.com/iyear/tdl/pkg/tplfunc"
	"github.com/iyear/tdl/pkg/utils"
)

func New(ctx context.Context, opts *Options) (*Iter, error) {
	tpl, err := template.New("dl").
		Funcs(tplfunc.FuncMap(tplfunc.All...)).
		Parse(opts.Template)
	if err != nil {
		return nil, err
	}

	dialogs := collectDialogs(opts.Dialogs)
	// if msgs is empty, return error to avoid range out of index
	if len(dialogs) == 0 {
		return nil, fmt.Errorf("you must specify at least one message")
	}

	// include and exclude
	includeMap := filterMap(opts.Include, utils.FS.AddPrefixDot)
	excludeMap := filterMap(opts.Exclude, utils.FS.AddPrefixDot)

	// to keep fingerprint stable
	sortDialogs(dialogs, opts.Desc)

	manager := peers.Options{Storage: storage.NewPeers(opts.KV)}.Build(opts.Pool.Default(ctx))
	it := &Iter{
		pool:        opts.Pool,
		dialogs:     dialogs,
		include:     includeMap,
		exclude:     excludeMap,
		curi:        0,
		curj:        -1,
		preSum:      preSum(dialogs),
		finished:    make(map[int]struct{}),
		template:    tpl,
		manager:     manager,
		fingerprint: fingerprint(dialogs),
	}

	return it, nil
}

func (iter *Iter) Next(ctx context.Context) (*downloader.Item, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	iter.mu.Lock()
	iter.curj++
	if iter.curj >= len(iter.dialogs[iter.curi].Messages) {
		if iter.curi++; iter.curi >= len(iter.dialogs) {
			return nil, errors.New("no more items")
		}
		iter.curj = 0
	}
	i, j := iter.curi, iter.curj
	iter.mu.Unlock()

	// check if finished
	if _, ok := iter.finished[iter.ij2n(i, j)]; ok {
		return nil, downloader.ErrSkip
	}

	return iter.item(ctx, i, j)
}

func (iter *Iter) item(ctx context.Context, i, j int) (*downloader.Item, error) {
	peer, msg := iter.dialogs[i].Peer, iter.dialogs[i].Messages[j]

	it := query.Messages(iter.pool.Default(ctx)).
		GetHistory(peer).OffsetID(msg + 1).
		BatchSize(1).Iter()
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
		return nil, fmt.Errorf("the message %d/%d may be deleted", id, msg)
	}

	item, ok := tmedia.GetMedia(message)
	if !ok {
		return nil, fmt.Errorf("can not get media from %d/%d message",
			id, message.ID)
	}

	// process include and exclude
	ext := filepath.Ext(item.Name)
	if len(iter.include) > 0 {
		if _, ok = iter.include[ext]; !ok {
			return nil, downloader.ErrSkip
		}
	}
	if len(iter.exclude) > 0 {
		if _, ok = iter.exclude[ext]; ok {
			return nil, downloader.ErrSkip
		}
	}

	buf := bytes.Buffer{}
	err := iter.template.Execute(&buf, &fileTemplate{
		DialogID:     id,
		MessageID:    message.ID,
		MessageDate:  int64(message.Date),
		FileName:     item.Name,
		FileSize:     utils.Byte.FormatBinaryBytes(item.Size),
		DownloadDate: time.Now().Unix(),
	})
	if err != nil {
		return nil, err
	}
	item.Name = buf.String()

	item.ID = iter.ij2n(i, j)

	return item, nil
}

func (iter *Iter) Finish(_ context.Context, id int) error {
	iter.mu.Lock()
	defer iter.mu.Unlock()

	iter.finished[id] = struct{}{}
	return nil
}

func (iter *Iter) Total(_ context.Context) int {
	iter.mu.Lock()
	defer iter.mu.Unlock()

	total := 0
	for _, m := range iter.dialogs {
		total += len(m.Messages)
	}
	return total
}

func (iter *Iter) ij2n(i, j int) int {
	return iter.preSum[i] + j
}

func (iter *Iter) SetFinished(finished map[int]struct{}) {
	iter.mu.Lock()
	defer iter.mu.Unlock()

	iter.finished = finished
}

func (iter *Iter) Finished() map[int]struct{} {
	iter.mu.Lock()
	defer iter.mu.Unlock()

	return iter.finished
}

func (iter *Iter) Fingerprint() string {
	return iter.fingerprint
}
