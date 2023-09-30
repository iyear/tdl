package utils

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/tg"
)

type telegram struct{}

var Telegram telegram

// ParseMessageLink return dialog id, msg id, error
func (t telegram) ParseMessageLink(ctx context.Context, manager *peers.Manager, s string) (peers.Peer, int, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, 0, err
	}

	paths := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")

	from, msg := "", ""
	switch len(paths) {
	case 2:
		// https://t.me/telegram/193

		from = paths[0]
		msg = paths[1]
	case 3:
		// https://t.me/c/1697797156/151
		// https://t.me/iFreeKnow/45662/55005

		if paths[0] == "c" {
			from = paths[1]
			msg = paths[2]
			break
		}

		// "45662" means topic id, we don't need it
		from = paths[0]
		msg = paths[2]
	case 4:
		// https://t.me/c/1492447836/251015/251021

		if paths[0] != "c" {
			return nil, 0, fmt.Errorf("invalid message link")
		}

		// "251015" means topic id, we don't need it
		from = paths[1]
		msg = paths[3]
	default:
		return nil, 0, fmt.Errorf("invalid message link: %s", s)
	}

	ch, err := t.GetInputPeer(ctx, manager, from)
	if err != nil {
		return nil, 0, err
	}

	msgid, err := strconv.Atoi(msg)
	if err != nil {
		return nil, 0, err
	}

	return ch, msgid, nil
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
