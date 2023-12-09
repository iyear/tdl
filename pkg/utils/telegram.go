package utils

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/tg"
)

type telegram struct{}

var Telegram telegram

// ParseMessageLink return dialog id, msg id, error
func (t telegram) ParseMessageLink(ctx context.Context, manager *peers.Manager, s string) (peers.Peer, int, error) {
	parse := func(from, msg string) (peers.Peer, int, error) {
		ch, err := t.GetInputPeer(ctx, manager, from)
		if err != nil {
			return nil, 0, errors.Wrap(err, "input peer")
		}

		m, err := strconv.Atoi(msg)
		if err != nil {
			return nil, 0, errors.Wrap(err, "parse message id")
		}

		return ch, m, nil
	}

	u, err := url.Parse(s)
	if err != nil {
		return nil, 0, err
	}

	paths := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")

	// https://t.me/opencfdchannel/4434?comment=360409
	if c := u.Query().Get("comment"); c != "" {
		peer, err := t.GetInputPeer(ctx, manager, paths[0])
		if err != nil {
			return nil, 0, errors.Wrap(err, "input peer")
		}

		ch, ok := peer.(peers.Channel)
		if !ok || !ch.IsBroadcast() {
			return nil, 0, errors.New("not channel")
		}

		raw, err := ch.FullRaw(ctx)
		if err != nil {
			return nil, 0, errors.Wrap(err, "full raw")
		}

		linked, ok := raw.GetLinkedChatID()
		if !ok {
			return nil, 0, errors.New("no linked chat")
		}

		return parse(strconv.FormatInt(linked, 10), c)
	}

	switch len(paths) {
	case 2:
		// https://t.me/telegram/193
		// https://t.me/myhostloc/1485524?thread=1485523
		return parse(paths[0], paths[1])
	case 3:
		// https://t.me/c/1697797156/151
		// https://t.me/iFreeKnow/45662/55005
		if paths[0] == "c" {
			return parse(paths[1], paths[2])
		}

		// "45662" means topic id, we don't need it
		return parse(paths[0], paths[2])
	case 4:
		// https://t.me/c/1492447836/251015/251021
		if paths[0] != "c" {
			return nil, 0, fmt.Errorf("invalid message link")
		}

		// "251015" means topic id, we don't need it
		return parse(paths[1], paths[3])
	default:
		return nil, 0, fmt.Errorf("invalid message link: %s", s)
	}
}

func (t telegram) GetInputPeer(ctx context.Context, manager *peers.Manager, from string) (peers.Peer, error) {
	id, err := strconv.ParseInt(from, 10, 64)
	if err != nil {
		// from is username
		p, err := manager.Resolve(ctx, from)
		if err != nil {
			return nil, err
		}

		return p, nil
	}

	var p peers.Peer
	if p, err = manager.ResolveChannelID(ctx, id); err == nil {
		return p, nil
	}
	if p, err = manager.ResolveUserID(ctx, id); err == nil {
		return p, nil
	}
	if p, err = manager.ResolveChatID(ctx, id); err == nil {
		return p, nil
	}

	return nil, fmt.Errorf("failed to get result from %d：%v", id, err)
}

func (t telegram) GetPeerID(peer tg.PeerClass) int64 {
	switch p := peer.(type) {
	case *tg.PeerUser:
		return p.UserID
	case *tg.PeerChat:
		return p.ChatID
	case *tg.PeerChannel:
		return p.ChannelID
	}
	return 0
}

func (t telegram) GetInputPeerID(peer tg.InputPeerClass) int64 {
	switch p := peer.(type) {
	case *tg.InputPeerUser:
		return p.UserID
	case *tg.InputPeerChat:
		return p.ChatID
	case *tg.InputPeerChannel:
		return p.ChannelID
	}

	return 0
}

func (t telegram) GetBlockedDialogs(ctx context.Context, client *tg.Client) (map[int64]struct{}, error) {
	blocks, err := query.GetBlocked(client).BatchSize(100).Collect(ctx)
	if err != nil {
		return nil, err
	}

	blockids := make(map[int64]struct{})
	for _, b := range blocks {
		blockids[t.GetPeerID(b.Contact.PeerID)] = struct{}{}
	}
	return blockids, nil
}

func (t telegram) FileExists(msg tg.MessageClass) bool {
	m, ok := msg.(*tg.Message)
	if !ok {
		return false
	}

	md, ok := m.GetMedia()
	if !ok {
		return false
	}

	switch md.(type) {
	case *tg.MessageMediaDocument, *tg.MessageMediaPhoto:
		return true
	default:
		return false
	}
}

func (t telegram) GetSingleMessage(ctx context.Context, c *tg.Client, peer tg.InputPeerClass, msg int) (*tg.Message, error) {
	it := query.Messages(c).
		GetHistory(peer).OffsetID(msg + 1).
		BatchSize(1).Iter()

	if !it.Next(ctx) {
		return nil, errors.Wrap(it.Err(), "get single message")
	}

	m, ok := it.Value().Msg.(*tg.Message)
	if !ok {
		return nil, errors.Errorf("invalid message %d", msg)
	}

	// check if message is deleted
	if m.GetID() != msg {
		return nil, errors.Errorf("the message %d/%d may be deleted", t.GetInputPeerID(peer), msg)
	}

	return m, nil
}

type Messages []*tg.Message

func (m Messages) Len() int {
	return len(m)
}

func (m Messages) Less(i, j int) bool {
	return m[i].ID < m[j].ID
}

func (m Messages) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func (t telegram) GetGroupedMessages(ctx context.Context, c *tg.Client, peer tg.InputPeerClass, msg *tg.Message) ([]*tg.Message, error) {
	group, ok := msg.GetGroupedID()
	if !ok {
		return nil, errors.New("not grouped message")
	}
	// https://telegram.org/blog/albums-saved-messages
	// Each album can include up to 10 photos or videos
	batchSize := 20

	it := query.Messages(c).GetHistory(peer).
		OffsetID(msg.ID + 11). // from latest to oldest
		BatchSize(batchSize).Iter()

	messages := make([]*tg.Message, 0, batchSize)
	for i := 0; it.Next(ctx) && i < batchSize; i++ {
		m, ok := it.Value().Msg.(*tg.Message)
		if !ok {
			continue
		}
		groupID, ok := m.GetGroupedID()
		if !ok {
			continue
		}
		if groupID != group {
			continue
		}

		messages = append(messages, m)
	}

	// reverse messages from oldest to latest, so we can forward them in order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

var threadsLevels = []struct {
	threads int
	size    int64
}{
	{1, 1 << 20},
	{2, 5 << 20},
	{4, 20 << 20},
	{8, 50 << 20},
}

func (t telegram) BestThreads(size int64, max int) int {
	// Get best threads num for download, based on file size
	for _, thread := range threadsLevels {
		if size < thread.size {
			return min(thread.threads, max)
		}
	}
	return max
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
