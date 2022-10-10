package utils

import (
	"context"
	"fmt"
	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/tg"
	"github.com/iyear/tdl/pkg/consts"
	"net/url"
	"strconv"
	"strings"
)

type telegram struct{}

var Telegram telegram

// ParseChannelMsgLink return dialog id, msg id, error
func (t telegram) ParseChannelMsgLink(ctx context.Context, manager *peers.Manager, s string) (peers.Peer, int, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, 0, err
	}

	paths := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")

	from, msg := "", ""
	switch len(paths) {
	case 2:
		from = paths[0]
		msg = paths[1]
	case 3:
		if paths[0] != "c" {
			return nil, 0, fmt.Errorf("invalid link path: %s", paths)
		}
		from = paths[1]
		msg = paths[2]
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

	if p, err := manager.ResolveChannelID(ctx, id); err == nil {
		return p, nil
	}

	if p, err := manager.ResolveUserID(ctx, id); err == nil {
		return p, nil
	}

	if p, err := manager.ResolveChatID(ctx, id); err == nil {
		return p, nil
	}

	return nil, fmt.Errorf("failed to get result from %d", id)
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

func (t telegram) GetName(first, last, username string) string {
	if name := first + " " + last; name != " " {
		return name
	}
	return username
}

func (t telegram) GetPeerName(id int64, e peer.Entities) string {
	if n, ok := e.Users()[id]; ok {
		return t.GetName(n.FirstName, n.LastName, n.Username)
	}

	if n, ok := e.Channels()[id]; ok {
		return n.Title
	}

	if n, ok := e.Chats()[id]; ok {
		return n.Title
	}

	return ""
}

func (t telegram) GetPeerType(id int64, e peer.Entities) string {
	if _, ok := e.User(id); ok {
		return consts.ChatPrivate
	}

	if n, ok := e.Channel(id); ok {
		if n.Megagroup || n.Gigagroup {
			return consts.ChatGroup
		}
		return consts.ChatChannel
	}

	if _, ok := e.Chat(id); ok {
		return consts.ChatGroup
	}

	return consts.ChatUnknown
}
