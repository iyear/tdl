package up

import (
	"fmt"
	"os"
	"sync"

	"github.com/fatih/color"
	"github.com/go-faster/errors"
	pw "github.com/jedib0t/go-pretty/v6/progress"
	"go.uber.org/multierr"

	"github.com/iyear/tdl/core/uploader"
	"github.com/iyear/tdl/pkg/prog"
	"github.com/iyear/tdl/pkg/utils"
)

type progress struct {
	pw       pw.Writer
	trackers sync.Map // map[*iterElem]*pw.Tracker
}

func newProgress(p pw.Writer) *progress {
	return &progress{
		pw: p,
	}
}

func (p *progress) OnAdd(elem uploader.Elem) {
	e := elem.(*iterElem)
	tracker := prog.AppendTracker(p.pw, utils.Byte.FormatBinaryBytes, p.elemString(e), e.file.Size())
	p.trackers.Store(e, tracker)
}

func (p *progress) OnUpload(elem uploader.Elem, state uploader.ProgressState) {
	e := elem.(*iterElem)
	tracker, ok := p.trackers.Load(e)
	if !ok {
		return
	}

	t := tracker.(*pw.Tracker)
	t.UpdateTotal(state.Total)
	if e.grouped && state.Uploaded >= state.Total {
		// Setting the full value would finish the tracker before Telegram confirms the send.
		t.UpdateMessage(p.elemString(e) + " (waiting for Telegram)")
		return
	}
	t.SetValue(state.Uploaded)
}

func (p *progress) OnDone(elem uploader.Elem, err error) {
	e := elem.(*iterElem)
	tracker, ok := p.trackers.LoadAndDelete(e)
	if !ok {
		return
	}
	t := tracker.(*pw.Tracker)
	if e.grouped {
		t.UpdateMessage(p.elemString(e))
	}

	if err = multierr.Append(err, e.close()); err != nil {
		p.fail(t, e, err)
		return
	}

	if e.remove {
		if err := os.Remove(e.file.File.Name()); err != nil {
			p.fail(t, e, errors.Wrap(err, "remove file"))
			return
		}
	}

	if e.grouped {
		t.SetValue(e.file.Size())
		t.MarkAsDone()
	}
}

func (p *progress) fail(t *pw.Tracker, elem *iterElem, err error) {
	p.pw.Log(color.RedString("%s error: %s", p.elemString(elem), err.Error()))
	t.MarkAsErrored()
}

func (p *progress) elemString(e *iterElem) string {
	return fmt.Sprintf("%s -> %s(%d)", e.file.File.Name(), e.to.VisibleName(), e.to.ID())
}
