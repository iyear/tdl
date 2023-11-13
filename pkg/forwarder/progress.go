package forwarder

import (
	"context"

	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
)

type Progress interface {
	OnAdd(peer peers.Peer, msg *tg.Message)
	OnClone(peer peers.Peer, msg *tg.Message, state ProgressState)
	OnDone(peer peers.Peer, msg *tg.Message, err error)
}

type ProgressState struct {
	Done  int64
	Total int64
}

type uploadProgress struct {
	peer     peers.Peer
	msg      *tg.Message
	progress Progress
}

func (p uploadProgress) Chunk(_ context.Context, state uploader.ProgressState) error {
	p.progress.OnClone(p.peer, p.msg, ProgressState{
		Done:  state.Uploaded,
		Total: state.Total,
	})
	return nil
}

type nopProgress struct{}

func (p nopProgress) Chunk(_ context.Context, _ uploader.ProgressState) error { return nil }
