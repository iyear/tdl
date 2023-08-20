package downloader

import (
	"os"
	"time"

	"github.com/jedib0t/go-pretty/v6/progress"
)

// writeAt wrapper for file to use progress bar
//
// do not need mutex because gotd has use syncio.WriteAt
type writeAt struct {
	f        *os.File
	tracker  *progress.Tracker
	partSize int
}

func newWriteAt(f *os.File, tracker *progress.Tracker, partSize int) *writeAt {
	return &writeAt{
		f:        f,
		tracker:  tracker,
		partSize: partSize,
	}
}

func (w *writeAt) WriteAt(p []byte, off int64) (int, error) {
	at, err := w.f.WriteAt(p, off)
	if err != nil {
		w.tracker.MarkAsErrored()
		return 0, err
	}

	// some small files may finish too fast, terminal history may not be overwritten
	// this is just a simple way to avoid the problem
	if at < w.partSize { //  last part(every file only exec once)
		time.Sleep(time.Millisecond * 200) // to ensure the progress render next time
	}
	w.tracker.Increment(int64(at))
	return at, nil
}
