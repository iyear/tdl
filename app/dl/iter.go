package dl

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/tg"
	"github.com/iyear/tdl/pkg/dcpool"
	"github.com/iyear/tdl/pkg/downloader"
	"github.com/iyear/tdl/pkg/kv"
	"github.com/iyear/tdl/pkg/storage"
	"github.com/iyear/tdl/pkg/utils"
	"path/filepath"
	"sort"
	"sync"
	"text/template"
	"time"
)

type iter struct {
	pool             dcpool.Pool
	dialogs          []*dialog
	include, exclude map[string]struct{}
	mu               sync.Mutex
	curi             int
	curj             int
	finished         map[int]struct{}
	template         *template.Template
	manager          *peers.Manager
	fingerprint      string
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

func newIter(pool dcpool.Pool, kvd kv.KV, tmpl string, include, exclude []string, items ...[]*dialog) (*iter, error) {
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

	// to keep fingerprint stable
	sortDialogs(mm)

	it := &iter{
		pool:        pool,
		dialogs:     mm,
		include:     includeMap,
		exclude:     excludeMap,
		curi:        0,
		curj:        -1,
		finished:    make(map[int]struct{}),
		template:    t,
		manager:     peers.Options{Storage: storage.NewPeers(kvd)}.Build(pool.Client(pool.Default())),
		fingerprint: fingerprint(mm),
	}

	return it, nil
}

func sortDialogs(dialogs []*dialog) {
	sort.Slice(dialogs, func(i, j int) bool {
		return utils.Telegram.GetInputPeerID(dialogs[i].peer) <
			utils.Telegram.GetInputPeerID(dialogs[j].peer) // increasing order
	})

	for _, m := range dialogs {
		sort.Slice(m.msgs, func(i, j int) bool {
			return m.msgs[i] > m.msgs[j] // decreasing order
		})
	}
}

func (iter *iter) Next(ctx context.Context) (*downloader.Item, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	iter.mu.Lock()
	iter.curj++
	if iter.curj >= len(iter.dialogs[iter.curi].msgs) {
		if iter.curi++; iter.curi >= len(iter.dialogs) {
			return nil, errors.New("no more items")
		}
		iter.curj = 0
	}
	iter.mu.Unlock()

	// check if finished
	if _, ok := iter.finished[iter.ij2n(iter.curi, iter.curj)]; ok {
		return nil, downloader.ErrSkip
	}

	return iter.item(ctx, iter.curi, iter.curj)
}

func (iter *iter) item(ctx context.Context, i, j int) (*downloader.Item, error) {
	peer, msg := iter.dialogs[i].peer, iter.dialogs[i].msgs[j]

	it := query.Messages(iter.pool.Client(iter.pool.Default())).
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
		return nil, fmt.Errorf("msg may be deleted, id: %d", msg)
	}

	media, ok := GetMedia(message)
	if !ok {
		return nil, fmt.Errorf("can not get media info: %d/%d",
			id, message.ID)
	}

	// process include and exclude
	ext := filepath.Ext(media.Name)
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
		FileName:     media.Name,
		FileSize:     utils.Byte.FormatBinaryBytes(media.Size),
		DownloadDate: time.Now().Unix(),
	})
	if err != nil {
		return nil, err
	}
	media.Name = buf.String()

	media.ID = iter.ij2n(i, j)

	return media, nil
}

func (iter *iter) setFinished(finished map[int]struct{}) {
	iter.mu.Lock()
	defer iter.mu.Unlock()

	iter.finished = finished
}

func (iter *iter) Finish(_ context.Context, id int) error {
	iter.mu.Lock()
	defer iter.mu.Unlock()

	iter.finished[id] = struct{}{}
	return nil
}

func (iter *iter) Total(_ context.Context) int {
	iter.mu.Lock()
	defer iter.mu.Unlock()

	total := 0
	for _, m := range iter.dialogs {
		total += len(m.msgs)
	}
	return total
}

func (iter *iter) ij2n(i, j int) int {
	n := 0
	for k := 0; k < i; k++ {
		n += len(iter.dialogs[k].msgs)
	}
	return n + j
}

func fingerprint(dialogs []*dialog) string {
	endian := binary.BigEndian
	buf, b := &bytes.Buffer{}, make([]byte, 8)
	for _, m := range dialogs {
		endian.PutUint64(b, uint64(utils.Telegram.GetInputPeerID(m.peer)))
		buf.Write(b)
		for _, msg := range m.msgs {
			endian.PutUint64(b, uint64(msg))
			buf.Write(b)
		}
	}

	return fmt.Sprintf("%x", sha256.Sum256(buf.Bytes()))
}
