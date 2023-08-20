package uploader

import (
	"context"

	"github.com/gotd/td/telegram/uploader"
	"github.com/jedib0t/go-pretty/v6/progress"
)

type _progress struct {
	tracker *progress.Tracker
}

func (p *_progress) Chunk(_ context.Context, state uploader.ProgressState) error {
	p.tracker.SetValue(state.Uploaded)
	return nil
}
