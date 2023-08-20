package downloader

import (
	"context"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/progress"

	"github.com/iyear/tdl/pkg/ps"
)

func (d *Downloader) renderPinned(ctx context.Context, pw progress.Writer) {
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
}
