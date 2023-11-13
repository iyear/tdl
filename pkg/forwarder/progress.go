package forwarder

import (
	"context"

	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
)

type Progress interface {
	OnAdd(meta *ProgressMeta)
	OnClone(meta *ProgressMeta, state ProgressState)
	OnDone(meta *ProgressMeta, err error)
}

type ProgressMeta struct {
	From peers.Peer
	Msg  *tg.Message
	To   peers.Peer
}

type ProgressState struct {
	Done  int64
	Total int64
}

type uploadProgress struct {
	meta     *ProgressMeta
	progress Progress
}

func (p uploadProgress) Chunk(_ context.Context, state uploader.ProgressState) error {
	p.progress.OnClone(p.meta, ProgressState{
		Done:  state.Uploaded,
		Total: state.Total,
	})
	return nil
}

type nopProgress struct{}

func (p nopProgress) Chunk(_ context.Context, _ uploader.ProgressState) error { return nil }
