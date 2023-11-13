package forward

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
	pw "github.com/jedib0t/go-pretty/v6/progress"
	"go.uber.org/zap"

	"github.com/iyear/tdl/pkg/forwarder"
	"github.com/iyear/tdl/pkg/prog"
	"github.com/iyear/tdl/pkg/utils"
)

type progress struct {
	pw       pw.Writer
	trackers map[[2]int64]*pw.Tracker
	log      *zap.Logger
}

func newProgress(p pw.Writer) *progress {
	return &progress{
		pw:       p,
		trackers: make(map[[2]int64]*pw.Tracker),
	}
}

func (p *progress) OnAdd(peer peers.Peer, msg *tg.Message) {
	tracker := prog.AppendTracker(p.pw, pw.FormatNumber, p.processMessage(peer, msg, false), 1)
	p.trackers[p.tuple(peer, msg)] = tracker
}

func (p *progress) OnClone(peer peers.Peer, msg *tg.Message, state forwarder.ProgressState) {
	tracker, ok := p.trackers[p.tuple(peer, msg)]
	if !ok {
		return
	}

	// display re-upload transfer info
	tracker.Units.Formatter = utils.Byte.FormatBinaryBytes
	tracker.UpdateMessage(p.processMessage(peer, msg, true))
	tracker.UpdateTotal(state.Total)
	tracker.SetValue(state.Done)
}

func (p *progress) OnDone(peer peers.Peer, msg *tg.Message, err error) {
	tracker, ok := p.trackers[p.tuple(peer, msg)]
	if !ok {
		return
	}

	if err != nil {
		p.pw.Log(color.RedString("%d-%d error: %s", peer.ID(), msg.ID, err.Error()))
		tracker.MarkAsErrored()
		return
	}

	tracker.Increment(1)
	tracker.MarkAsDone()
}

func (p *progress) tuple(peer peers.Peer, msg *tg.Message) [2]int64 {
	return [2]int64{peer.ID(), int64(msg.ID)}
}

func (p *progress) processMessage(peer peers.Peer, msg *tg.Message, clone bool) string {
	b := &strings.Builder{}

	// TODO(iyear): display visible name which should be cut to 20 chars
	b.WriteString(fmt.Sprintf("%d-%d", peer.ID(), msg.ID))
	if clone {
		b.WriteString(" [clone]")
	}

	return b.String()
}
