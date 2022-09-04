package downloader

import (
	"github.com/jedib0t/go-pretty/v6/progress"
	"os"
	"sync"
)

// writeAt wrapper for file to use progress bar
type writeAt struct {
	mu      sync.Mutex
	f       *os.File
	tracker *progress.Tracker
}

func (w *writeAt) WriteAt(p []byte, off int64) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	at, err := w.f.WriteAt(p, off)
	if err != nil {
		w.tracker.MarkAsErrored()
		return 0, err
	}
	w.tracker.Increment(int64(at))
	return at, nil
}
