package forward

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	pw "github.com/jedib0t/go-pretty/v6/progress"
	"github.com/mattn/go-runewidth"

	"github.com/iyear/tdl/pkg/forwarder"
	"github.com/iyear/tdl/pkg/prog"
	"github.com/iyear/tdl/pkg/utils"
)

type progress struct {
	pw       pw.Writer
	trackers map[[3]int64]*pw.Tracker
	meta     map[int64]string
}

func newProgress(p pw.Writer) *progress {
	return &progress{
		pw:       p,
		trackers: make(map[[3]int64]*pw.Tracker),
		meta:     make(map[int64]string),
	}
}

func (p *progress) OnAdd(meta *forwarder.ProgressMeta) {
	tracker := prog.AppendTracker(p.pw, pw.FormatNumber, p.processMessage(meta, false), 1)
	p.trackers[p.tuple(meta)] = tracker
}

func (p *progress) OnClone(meta *forwarder.ProgressMeta, state forwarder.ProgressState) {
	tracker, ok := p.trackers[p.tuple(meta)]
	if !ok {
		return
	}

	// display re-upload transfer info
	tracker.Units.Formatter = utils.Byte.FormatBinaryBytes
	tracker.UpdateMessage(p.processMessage(meta, true))
	tracker.UpdateTotal(state.Total)
	tracker.SetValue(state.Done)
}

func (p *progress) OnDone(meta *forwarder.ProgressMeta, err error) {
	tracker, ok := p.trackers[p.tuple(meta)]
	if !ok {
		return
	}

	if err != nil {
		p.pw.Log(color.RedString("%s error: %s", p.metaString(meta), err.Error()))
		tracker.MarkAsErrored()
		return
	}

	tracker.Increment(1)
	tracker.MarkAsDone()
}

func (p *progress) tuple(meta *forwarder.ProgressMeta) [3]int64 {
	return [3]int64{meta.From.ID(), int64(meta.Msg.ID), meta.To.ID()}
}

func (p *progress) processMessage(meta *forwarder.ProgressMeta, clone bool) string {
	b := &strings.Builder{}

	b.WriteString(p.metaString(meta))
	if clone {
		b.WriteString(" [clone]")
	}

	return b.String()
}

func (p *progress) metaString(meta *forwarder.ProgressMeta) string {
	// TODO(iyear): better responsive name
	if _, ok := p.meta[meta.From.ID()]; !ok {
		p.meta[meta.From.ID()] = runewidth.Truncate(meta.From.VisibleName(), 15, "...")
	}
	if _, ok := p.meta[meta.To.ID()]; !ok {
		p.meta[meta.To.ID()] = runewidth.Truncate(meta.To.VisibleName(), 15, "...")
	}

	return fmt.Sprintf("%s(%d):%d -> %s(%d)", p.meta[meta.From.ID()], meta.From.ID(), meta.Msg.ID, p.meta[meta.To.ID()], meta.To.ID())
}
