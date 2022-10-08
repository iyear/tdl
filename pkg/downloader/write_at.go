package downloader

import (
	"github.com/jedib0t/go-pretty/v6/progress"
	"os"
)

// writeAt wrapper for file to use progress bar
//
// do not need mutex because gotd has use syncio.WriteAt
type writeAt struct {
	f       *os.File
	tracker *progress.Tracker
}

func (w *writeAt) WriteAt(p []byte, off int64) (int, error) {
	at, err := w.f.WriteAt(p, off)
	if err != nil {
		w.tracker.MarkAsErrored()
		return 0, err
	}
	w.tracker.Increment(int64(at))
	return at, nil
}
