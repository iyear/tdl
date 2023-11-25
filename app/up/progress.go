package up

import (
	"fmt"

	"github.com/fatih/color"
	pw "github.com/jedib0t/go-pretty/v6/progress"

	"github.com/iyear/tdl/pkg/prog"
	"github.com/iyear/tdl/pkg/uploader"
	"github.com/iyear/tdl/pkg/utils"
)

type progress struct {
	pw       pw.Writer
	trackers map[tuple]*pw.Tracker
}

type tuple struct {
	name string
	to   int64
}

func newProgress(p pw.Writer) *progress {
	return &progress{
		pw:       p,
		trackers: make(map[tuple]*pw.Tracker),
	}
}

func (p *progress) OnAdd(elem *uploader.Elem) {
	tracker := prog.AppendTracker(p.pw, utils.Byte.FormatBinaryBytes, p.processMessage(elem), elem.Size)
	p.trackers[p.tuple(elem)] = tracker
}

func (p *progress) OnUpload(elem *uploader.Elem, state uploader.ProgressState) {
	tracker, ok := p.trackers[p.tuple(elem)]
	if !ok {
		return
	}

	tracker.UpdateTotal(state.Total)
	tracker.SetValue(state.Uploaded)
}

func (p *progress) OnDone(elem *uploader.Elem, err error) {
	tracker, ok := p.trackers[p.tuple(elem)]
	if !ok {
		return
	}

	if err != nil {
		p.pw.Log(color.RedString("%s error: %s", p.elemString(elem), err.Error()))
		tracker.MarkAsErrored()
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
