package prog

import (
	"context"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/progress"

	"github.com/iyear/tdl/pkg/ps"
)

func AppendTracker(pw progress.Writer, formatter progress.UnitsFormatter, message string, total int64) *progress.Tracker {
	units := progress.UnitsBytes
	units.Formatter = formatter

	tracker := progress.Tracker{
		Message: message,
		Total:   total,
		Units:   units,
	}

	pw.AppendTracker(&tracker)

	return &tracker
}

// EnablePS enables pinned messages with ps info: cpu, memory, goroutines.
// It returns a stop function to clear the pinned message and stop updates.
func EnablePS(ctx context.Context, pw progress.Writer) func() {
	ctx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})

	go func() {
		defer close(done)
		f := func() { pw.SetPinnedMessages(strings.Join(ps.Humanize(ctx), " ")) }
		f()

		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				pw.SetPinnedMessages()
				return
			case <-ticker.C:
				f()
			}
		}
	}()

	return func() {
		cancel()
		<-done
	}
}
