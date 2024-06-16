package dl

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/gabriel-vasile/mimetype"
	"github.com/go-faster/errors"
	pw "github.com/jedib0t/go-pretty/v6/progress"

	"github.com/iyear/tdl/core/downloader"
	"github.com/iyear/tdl/core/util/fsutil"
	"github.com/iyear/tdl/pkg/prog"
	"github.com/iyear/tdl/pkg/utils"
)

type progress struct {
	pw       pw.Writer
	trackers *sync.Map // map[ID]*pw.Tracker
	opts     Options

	it *iter
}

func newProgress(p pw.Writer, it *iter, opts Options) *progress {
	return &progress{
		pw:       p,
		trackers: &sync.Map{},
		opts:     opts,
		it:       it,
	}
}

func (p *progress) OnAdd(elem downloader.Elem) {
	tracker := prog.AppendTracker(p.pw, utils.Byte.FormatBinaryBytes, p.processMessage(elem), elem.File().Size())
	p.trackers.Store(elem.(*iterElem).id, tracker)
}

func (p *progress) OnDownload(elem downloader.Elem, state downloader.ProgressState) {
	tracker, ok := p.trackers.Load(elem.(*iterElem).id)
	if !ok {
		return
	}

	t := tracker.(*pw.Tracker)
	t.UpdateTotal(state.Total)
	t.SetValue(state.Downloaded)
}

func (p *progress) OnDone(elem downloader.Elem, err error) {
	e := elem.(*iterElem)

	tracker, ok := p.trackers.Load(e.id)
	if !ok {
		return
	}
	t := tracker.(*pw.Tracker)

	if err := e.to.Close(); err != nil {
		p.fail(t, elem, errors.Wrap(err, "close file"))
		return
	}

	if err != nil {
		if !errors.Is(err, context.Canceled) { // don't report user cancel
			p.fail(t, elem, errors.Wrap(err, "progress"))
		}
		_ = os.Remove(e.to.Name()) // just try to remove temp file, ignore error
		return
	}

	p.it.Finish(e.id)

	if err := p.donePost(e); err != nil {
		p.fail(t, elem, errors.Wrap(err, "post file"))
		return
	}
}

func (p *progress) donePost(elem *iterElem) error {
	newfile := strings.TrimSuffix(filepath.Base(elem.to.Name()), tempExt)

	if p.opts.RewriteExt {
		mime, err := mimetype.DetectFile(elem.to.Name())
		if err != nil {
			return errors.Wrap(err, "detect mime")
		}
		ext := mime.Extension()
		if ext != "" && (filepath.Ext(newfile) != ext) {
			newfile = fsutil.GetNameWithoutExt(newfile) + ext
		}
	}

	if err := os.Rename(elem.to.Name(), filepath.Join(filepath.Dir(elem.to.Name()), newfile)); err != nil {
		return errors.Wrap(err, "rename file")
	}

	return nil
}

func (p *progress) fail(t *pw.Tracker, elem downloader.Elem, err error) {
	p.pw.Log(color.RedString("%s error: %s", p.elemString(elem), err.Error()))
	t.MarkAsErrored()
}

func (p *progress) processMessage(elem downloader.Elem) string {
	return p.elemString(elem)
}

func (p *progress) elemString(elem downloader.Elem) string {
	e := elem.(*iterElem)
	return fmt.Sprintf("%s(%d):%d -> %s",
		e.from.VisibleName(),
		e.from.ID(),
		e.fromMsg.ID,
		strings.TrimSuffix(e.to.Name(), tempExt))
}
