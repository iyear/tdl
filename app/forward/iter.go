package forward

import (
	"context"

	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram/peers"

	"github.com/iyear/tdl/pkg/dcpool"
	"github.com/iyear/tdl/pkg/forwarder"
	"github.com/iyear/tdl/pkg/tmessage"
	"github.com/iyear/tdl/pkg/utils"
)

type iter struct {
	manager *peers.Manager
	pool    dcpool.Pool
	to      peers.Peer
	dialogs []*tmessage.Dialog
	i, j    int
	elem    *forwarder.Elem
	err     error
}

func newIter(manager *peers.Manager, pool dcpool.Pool, to peers.Peer, dialogs []*tmessage.Dialog) *iter {
	return &iter{
		manager: manager,
		pool:    pool,
		to:      to,
		dialogs: dialogs,
	}
}

func (i *iter) Next(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		i.err = ctx.Err()
		return false
	default:
	}

	if i.i >= len(i.dialogs) || i.j >= len(i.dialogs[i.i].Messages) || i.err != nil {
		return false
	}

	p, m := i.dialogs[i.i].Peer, i.dialogs[i.i].Messages[i.j]

	if i.j++; i.j >= len(i.dialogs[i.i].Messages) {
		i.i++
		i.j = 0
	}

	peer, err := i.manager.FromInputPeer(ctx, p)
	if err != nil {
		i.err = errors.Wrap(err, "get peer")
		return false
	}

	msg, err := utils.Telegram.GetSingleMessage(ctx, i.pool.Default(ctx), peer.InputPeer(), m)
	if err != nil {
		i.err = errors.Wrapf(err, "get message %d", msg.ID)
		return false
	}

	i.elem = &forwarder.Elem{
		From: peer,
		To:   i.to,
		Msg:  msg,
	}

	return true
}

func (i *iter) Value() *forwarder.Elem {
	return i.elem
}

func (i *iter) Err() error {
	return i.err
}
