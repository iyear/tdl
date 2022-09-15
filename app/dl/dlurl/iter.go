package dlurl

import (
	"context"
	"fmt"
	"github.com/gotd/td/tg"
	"github.com/iyear/tdl/app/dl"
	"github.com/iyear/tdl/pkg/downloader"
)

type iter struct {
	client *tg.Client
	msgs   []*msg
	cur    int
}

type msg struct {
	ch  *tg.InputChannel
	msg int
}

func newIter(client *tg.Client, msgs []*msg) *iter {
	return &iter{
		client: client,
		msgs:   msgs,
		cur:    -1,
	}
}

func (i *iter) Next(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	default:
	}

	i.cur++

	if i.cur == len(i.msgs) {
		return false
	}

	return true
}

func (i *iter) Value(ctx context.Context) (*downloader.Item, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	cur := i.msgs[i.cur]

	msgs, err := i.client.ChannelsGetMessages(ctx, &tg.ChannelsGetMessagesRequest{
		Channel: cur.ch,
		ID:      []tg.InputMessageClass{&tg.InputMessageID{ID: cur.msg}},
	})
	if err != nil {
		return nil, err
	}

	m, ok := msgs.(*tg.MessagesChannelMessages)
	if !ok {
		return nil, fmt.Errorf("msg is not *tg.MessagesChannelMessages")
	}

	if len(m.Messages) != 1 {
		return nil, fmt.Errorf("len(msg) is not 1")
	}

	item, ok := dl.GetMedia(m.Messages[0])
	if !ok {
		return nil, fmt.Errorf("can not get media info")
	}

	item.Name = fmt.Sprintf("%d_%d_%s", cur.ch.ChannelID, cur.msg, item.Name)

	return item, nil
}

func (i *iter) Total(_ context.Context) int {
	return len(i.msgs)
}
