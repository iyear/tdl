package dl

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/tg"
	"github.com/iyear/tdl/pkg/dcpool"
	"github.com/iyear/tdl/pkg/downloader"
	"github.com/iyear/tdl/pkg/kv"
	"github.com/iyear/tdl/pkg/storage"
	"github.com/iyear/tdl/pkg/utils"
)

type iter struct {
	pool             dcpool.Pool
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

func checktgResumes(d []*dialog) {

	for _, i := range d {

		//strings.Fields()
		//c := reflect.ValueOf(i.peer)
		//c.CanInterface()
		//c.FieldByName("ChannelID")
		//i.peer.ChannelID
		//b := i.peer.GetChannelID(bool)

		// Temporary Hacky way to get the channelID since calling GetChannelID() only works runtime with dlv
		channelid := strings.Fields(i.peer.String())[0]
		_, channelid, _ = strings.Cut(channelid, ":")
		println("ChannelID ->", channelid)

		fmt.Println("DEBUG -> i.msgs at start of function ", i.msgs)

		resfile := fmt.Sprintf("%s%s", channelid, downloader.ResExt)

		f, err := os.Open(resfile)
		if err != nil {
			fmt.Println(err.Error() + "while opening: " + resfile)
			// TODO: Right now we continue because the first loop but we should crash once we figure out the new signature
			continue
		}
		defer f.Close()

		// Grab completed from file
		var completes []int
		r := bufio.NewReader(f)
		for {
			line, err := r.ReadString(10) // 0x0A separator = newline
			if (err != nil) && (err.Error() == "EOF") {
				break // just end here
			} else if err != nil {
				fmt.Println(err.Error() + "while processing: " + resfile)
			}

			line = strings.TrimSpace(line)
			line = strings.TrimRight(line, "\n")
			var c int
			c, _ = strconv.Atoi(line)
			completes = append(completes, c)
		}
		fmt.Println("DEBUG -> TGResume File Completed ", completes)

		// Create new temporary slice to filter out completes
		var tempslice []int
		for _, msg := range i.msgs {

			if !isCompleted(msg, completes) {
				tempslice = append(tempslice, msg)
			}
		}

		// Change the original slice
		i.msgs = tempslice
		fmt.Println("DEBUG -> i.msgs at end of function ", i.msgs)
	}
	return
}

func isCompleted(m int, completed []int) bool {

	for _, e := range completed {
		if m == e {
			return true
		}
	}
	return false

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

	checktgResumes(mm)
	fmt.Println("DEBUG -> mm[0] outside of function check", mm[0].msgs)

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
		pool:     pool,
		dialogs:  mm,
		include:  includeMap,
		exclude:  excludeMap,
		curi:     0,
		curj:     -1,
		template: t,
		manager:  peers.Options{Storage: storage.NewPeers(kvd)}.Build(pool.Client(pool.Default())),
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
	it := query.Messages(i.pool.Client(i.pool.Default())).
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

	// Check If message has been successfully downloaded and skip
	//if CheckResumableForCompleted(message.ID) {
	//	return nil, fmt.Errorf("msg already completed downloaded, id: %d", msg)
	//  return nil, downloader.ErrSkip   // Prob choose this one as it wont log according to downloader.goRu
	//}

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
	media.ChatID = id
	media.MsgID = message.ID

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
