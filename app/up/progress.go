package up

import (
	"fmt"
	"sync"

	"github.com/fatih/color"
	pw "github.com/jedib0t/go-pretty/v6/progress"

	"github.com/iyear/tdl/pkg/prog"
	"github.com/iyear/tdl/pkg/uploader"
	"github.com/iyear/tdl/pkg/utils"
)

type progress struct {
	pw       pw.Writer
	trackers *sync.Map // map[tuple]*pw.Tracker
}

type tuple struct {
	name string
	to   int64
}

func newProgress(p pw.Writer) *progress {
	return &progress{
		pw:       p,
		trackers: &sync.Map{},
	}
}

func (p *progress) OnAdd(elem *uploader.Elem) {
	tracker := prog.AppendTracker(p.pw, utils.Byte.FormatBinaryBytes, p.processMessage(elem), elem.Size)
	p.trackers.Store(p.tuple(elem), tracker)
}

func (p *progress) OnUpload(elem *uploader.Elem, state uploader.ProgressState) {
	tracker, ok := p.trackers.Load(p.tuple(elem))
	if !ok {
		return
	}

	t := tracker.(*pw.Tracker)
	t.UpdateTotal(state.Total)
	t.SetValue(state.Uploaded)
}

func (p *progress) OnDone(elem *uploader.Elem, err error) {
	tracker, ok := p.trackers.Load(p.tuple(elem))
	if !ok {
		return
	}

	t := tracker.(*pw.Tracker)
	if err != nil {
		p.pw.Log(color.RedString("%s error: %s", p.elemString(elem), err.Error()))
		t.MarkAsErrored()
		return
	}
}

func (p *progress) tuple(elem *uploader.Elem) tuple {
	return tuple{elem.Name, elem.To.ID()}
}

func (p *progress) processMessage(elem *uploader.Elem) string {
	return p.elemString(elem)
}

func (p *progress) elemString(elem *uploader.Elem) string {
	return fmt.Sprintf("%s -> %s(%d)", elem.Name, elem.To.VisibleName(), elem.To.ID())
}
