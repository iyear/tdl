package downloader

import (
	"github.com/jedib0t/go-pretty/v6/progress"
	"os"
	"time"
)

// writeAt wrapper for file to use progress bar
//
// do not need mutex because gotd has use syncio.WriteAt
type writeAt struct {
	f       *os.File
	tracker *progress.Tracker
}

func newWriteAt(f *os.File, tracker *progress.Tracker) *writeAt {
	return &writeAt{
		f:       f,
		tracker: tracker,
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
	if w.tracker.Value()+int64(at) >= w.tracker.Total {
		time.Sleep(time.Millisecond * 200) // to ensure the progress render next time
	}
	w.tracker.Increment(int64(at))
	return at, nil
}
