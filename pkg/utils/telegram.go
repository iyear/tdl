package utils

import (
	"context"
	"fmt"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
	"net/url"
	"strconv"
	"strings"
)

type telegram struct{}

var Telegram telegram

// ParseChannelMsgLink return dialog id, msg id, error
func (t telegram) ParseChannelMsgLink(ctx context.Context, manager *peers.Manager, s string) (*tg.InputChannel, int, error) {
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

	ch, err := t.GetInputChannel(ctx, manager, from)
	if err != nil {
		return nil, 0, err
	}

	msgid, err := strconv.Atoi(msg)
	if err != nil {
		return nil, 0, err
	}

	return ch, msgid, nil
}

func (t telegram) GetInputChannel(ctx context.Context, manager *peers.Manager, from string) (*tg.InputChannel, error) {
	id, err := strconv.ParseInt(from, 10, 64)
	if err != nil {
		// from is username
		peer, err := manager.ResolveDomain(ctx, from)
		if err != nil {
			return nil, err
		}

		ch, ok := peer.InputPeer().(*tg.InputPeerChannel)
		if !ok {
			return nil, err
		}

		return &tg.InputChannel{ChannelID: ch.ChannelID, AccessHash: ch.AccessHash}, nil
	}

	ch, err := manager.ResolveChannelID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &tg.InputChannel{ChannelID: ch.Raw().ID, AccessHash: ch.Raw().AccessHash}, nil
}
