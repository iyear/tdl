package forward

import (
	"context"

	"github.com/antonmedv/expr/vm"
	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"

	"github.com/iyear/tdl/pkg/dcpool"
	"github.com/iyear/tdl/pkg/forwarder"
	"github.com/iyear/tdl/pkg/texpr"
	"github.com/iyear/tdl/pkg/tmessage"
	"github.com/iyear/tdl/pkg/utils"
)

type iterOptions struct {
	manager *peers.Manager
	pool    dcpool.Pool
	to      *vm.Program
	dialogs []*tmessage.Dialog
	mode    forwarder.Mode
	silent  bool
	dryRun  bool
}

type iter struct {
	opts iterOptions

	i, j int
	elem forwarder.Elem
	err  error
}

type env struct {
	From struct {
		ID          int64  `comment:"ID of dialog"`
		Username    string `comment:"Username of dialog"`
		VisibleName string `comment:"Title of channel and group, first and last name of user"`
	}
	Message texpr.EnvMessage
}

func exprEnv(from peers.Peer, msg *tg.Message) env {
	e := env{}

	if from != nil {
		e.From.ID = from.ID()
		e.From.Username, _ = from.Username()
		e.From.VisibleName = from.VisibleName()
	}

	if msg != nil {
		e.Message = texpr.ConvertEnvMessage(msg)
	}

	return e
}

func newIter(opts iterOptions) *iter {
	return &iter{
		opts: opts,

		i:    0,
		j:    0,
		elem: nil,
		err:  nil,
	}
}

func (i *iter) Next(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		i.err = ctx.Err()
		return false
	default:
	}

	// end of iteration or error occurred
	if i.i >= len(i.opts.dialogs) || i.err != nil {
		return false
	}

	p, m := i.opts.dialogs[i.i].Peer, i.opts.dialogs[i.i].Messages[i.j]

	if i.j++; i.j >= len(i.opts.dialogs[i.i].Messages) {
		i.i++
		i.j = 0
	}

	from, err := i.opts.manager.FromInputPeer(ctx, p)
	if err != nil {
		i.err = errors.Wrap(err, "get from peer")
		return false
	}

	msg, err := utils.Telegram.GetSingleMessage(ctx, i.opts.pool.Default(ctx), from.InputPeer(), m)
	if err != nil {
		i.err = errors.Wrapf(err, "get message: %d", m)
		return false
	}

	// message routing
	result, err := texpr.Run(i.opts.to, exprEnv(from, msg))
	if err != nil {
		i.err = errors.Wrap(err, "message routing")
		return false
	}
	destPeer, ok := result.(string)
	if !ok {
		i.err = errors.Errorf("message router must return string: %T", result)
		return false
	}

	var to peers.Peer
	if destPeer == "" { // self
		to, err = i.opts.manager.Self(ctx)
	} else {
		to, err = utils.Telegram.GetInputPeer(ctx, i.opts.manager, destPeer)
	}

	if err != nil {
		i.err = errors.Wrapf(err, "resolve dest peer: %s", destPeer)
		return false
	}

	i.elem = &iterElem{
		from: from,
		msg:  msg,
		to:   to,
		opts: i.opts,
	}

	return true
}

func (i *iter) Value() forwarder.Elem {
	return i.elem
}

func (i *iter) Err() error {
	return i.err
}
