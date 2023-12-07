package forwarder

import (
	"context"

	"github.com/gotd/td/telegram/uploader"
)

type Progress interface {
	OnAdd(elem Elem)
	OnClone(elem Elem, state ProgressState)
	OnDone(elem Elem, err error)
}

type ProgressState struct {
	Done  int64
	Total int64
}

type uploadProgress struct {
	elem     Elem
	progress Progress
}

func (p uploadProgress) Chunk(_ context.Context, state uploader.ProgressState) error {
	p.progress.OnClone(p.elem, ProgressState{
		Done:  state.Uploaded,
		Total: state.Total,
	})
	return nil
}

type nopProgress struct{}

func (p nopProgress) Chunk(_ context.Context, _ uploader.ProgressState) error { return nil }
