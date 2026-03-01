package dl

import (
	"context"
	stdErrors "errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

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

	next downloader.Progress // external progress listener
}

func newProgress(p pw.Writer, it *iter, opts Options) *progress {
	return &progress{
		pw:       p,
		trackers: &sync.Map{},
		opts:     opts,
		it:       it,
		next:     opts.ExternalProgress,
	}
}

func (p *progress) OnAdd(elem downloader.Elem) {
	if p.next != nil {
		p.next.OnAdd(elem)
	}

	if p.opts.Silent {
		return
	}

	tracker := prog.AppendTracker(p.pw, utils.Byte.FormatBinaryBytes, p.processMessage(elem), elem.File().Size())
	p.trackers.Store(elem.(*iterElem).id, tracker)
}

func (p *progress) OnDownload(elem downloader.Elem, state downloader.ProgressState) {
	if p.next != nil {
		p.next.OnDownload(elem, state)
	}

	if p.opts.Silent {
		return
	}

	tracker, ok := p.trackers.Load(elem.(*iterElem).id)
	if !ok {
		return
	}

	t := tracker.(*pw.Tracker)
	t.UpdateTotal(state.Total)
	t.SetValue(state.Downloaded)
}

func (p *progress) OnDone(elem downloader.Elem, err error) {
	if p.next != nil {
		p.next.OnDone(elem, err)
	}

	e := elem.(*iterElem)

	// Always cleanup file handles regardless of Silent mode or UI
	// ... (rest of the logic remains same until tracker update)

	// Optional: ensure any buffered data is flushed to disk before closing/renaming.
	_ = e.to.Sync()

	if err := e.to.Close(); err != nil {
		p.fail(elem, errors.Wrap(err, "close file"))
		return
	}

	if err != nil {
		if !errors.Is(err, context.Canceled) { // don't report user cancel
			p.fail(elem, errors.Wrap(err, "progress"))
		}
		_ = os.Remove(e.to.Name()) // just try to remove temp file, ignore error
		return
	}

	p.it.Finish(e.logicalPos)

	if err := p.donePost(e); err != nil {
		p.fail(elem, errors.Wrap(err, "post file"))
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

	newpath := filepath.Join(filepath.Dir(elem.to.Name()), newfile)

	// Windows can temporarily lock files (Defender/AV/Indexer/Explorer preview).
	// Retry rename to avoid failing the download at the final step.
	if err := renameWithRetry(elem.to.Name(), newpath); err != nil {
		return errors.Wrap(err, "rename file")
	}

	// Set file modification time to message date if available
	if elem.file.Date > 0 {
		fileTime := time.Unix(elem.file.Date, 0)
		if err := os.Chtimes(newpath, fileTime, fileTime); err != nil {
			return errors.Wrap(err, "set file time")
		}
	}

	return nil
}

func (p *progress) fail(elem downloader.Elem, err error) {
	// invoke next progress handler if it exists
	// we can report error via OnDone if not finished?
	// But OnDone is called with err already.
	// fail() acts as a helper in OnDone.
	// So p.next.OnDone is already called with err if we passed it.

	if p.opts.Silent {
		return
	}

	tracker, ok := p.trackers.Load(elem.(*iterElem).id)
	if !ok {
		return
	}
	t := tracker.(*pw.Tracker)

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

func renameWithRetry(oldpath, newpath string) error {
	const (
		// On some Windows machines (heavy AV/Defender, slow disks, etc.),
		// the temp file or destination can stay locked for quite a while
		// after we close our own handle. A small retry window (~9s) is
		// often not enough for large media files, which leads to
		// "post file: rename file" errors and forces a re-download.
		//
		// We therefore allow a much longer retry window here on Windows
		// (attempts * delay), while still bailing out quickly on other
		// platforms or non-lock related errors.
		attempts = 2000
		delay    = 100 * time.Millisecond
	)

	var err error
	for i := 0; i < attempts; i++ {
		err = os.Rename(oldpath, newpath)
		if err == nil {
			return nil
		}

		// Only retry transient Windows locking errors.
		if runtime.GOOS != "windows" || !isWindowsFileLockError(err) {
			return err
		}

		time.Sleep(delay)
	}
	return err
}

func isWindowsFileLockError(err error) bool {
	// Numeric errno values so this compiles cross-platform.
	// 5  = Access is denied
	// 32 = Sharing violation
	// 33 = Lock violation
	const (
		winAccessDenied     syscall.Errno = 5
		winSharingViolation syscall.Errno = 32
		winLockViolation    syscall.Errno = 33
	)

	for err != nil {
		if pe, ok := err.(*os.PathError); ok {
			err = pe.Err
			continue
		}

		if errno, ok := err.(syscall.Errno); ok {
			return errno == winAccessDenied ||
				errno == winSharingViolation ||
				errno == winLockViolation
		}

		err = stdErrors.Unwrap(err)
	}
	return false
}
