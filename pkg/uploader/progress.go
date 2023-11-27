package uploader

import (
	"context"

	"github.com/gotd/td/telegram/uploader"
)

type Progress interface {
	OnAdd(elem Elem)
	OnUpload(elem Elem, state ProgressState)
	OnDone(elem Elem, err error)
	// TODO: OnLog to log something that is not an error but should be sent to the user
}

type ProgressState struct {
	Uploaded int64
	Total    int64
}

type wrapProcess struct {
	elem    Elem
	process Progress
}

func (p *wrapProcess) Chunk(_ context.Context, state uploader.ProgressState) error {
	p.process.OnUpload(p.elem, ProgressState{
		Uploaded: state.Uploaded,
		Total:    state.Total,
	})
	return nil
}
