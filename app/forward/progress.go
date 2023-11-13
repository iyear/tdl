package forward

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	pw "github.com/jedib0t/go-pretty/v6/progress"

	"github.com/iyear/tdl/pkg/forwarder"
	"github.com/iyear/tdl/pkg/prog"
	"github.com/iyear/tdl/pkg/utils"
)

type progress struct {
	pw       pw.Writer
	trackers map[[3]int64]*pw.Tracker
}

func newProgress(p pw.Writer) *progress {
	return &progress{
		pw:       p,
		trackers: make(map[[3]int64]*pw.Tracker),
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
		p.pw.Log(color.RedString("%d-%d error: %s", meta.From.ID(), meta.Msg.ID, err.Error()))
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

	// TODO(iyear): display visible name which should be cut to 20 chars
	b.WriteString(fmt.Sprintf("%d-%d-%d", meta.From.ID(), meta.Msg.ID, meta.To.ID()))
	if clone {
		b.WriteString(" [clone]")
	}

	return b.String()
}
