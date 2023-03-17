package chat

import (
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/gotd/contrib/middleware/ratelimit"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/tg"
	"github.com/iyear/tdl/app/internal/tgc"
	"github.com/iyear/tdl/pkg/prog"
	"github.com/iyear/tdl/pkg/storage"
	"github.com/iyear/tdl/pkg/utils"
	"github.com/jedib0t/go-pretty/v6/progress"
	"golang.org/x/time/rate"
	"io"
	"os"
	"time"
)

const (
	rateInterval = 550 * time.Millisecond
	rateBucket   = 2
)

type ExportOptions struct {
	Chat   string
	From   int
	To     int
	Output string
	Time   bool
	Msg    bool
}

func Export(ctx context.Context, opts *ExportOptions) error {
	c, kvd, err := tgc.NoLogin(ctx, ratelimit.New(rate.Every(rateInterval), rateBucket))
	if err != nil {
		return err
	}

	return tgc.RunWithAuth(ctx, c, func(ctx context.Context) error {
		manager := peers.Options{Storage: storage.NewPeers(kvd)}.Build(c.API())
		peer, err := utils.Telegram.GetInputPeer(ctx, manager, opts.Chat)
		if err != nil {
			return err
		}

		color.Yellow("WARN: Export only generates minimal JSON for tdl download, not for backup.")
		color.Cyan("Occasional suspensions are due to Telegram rate limitations, please wait a moment.")
		fmt.Println()

		color.Blue("Indexing... [%s/%d]", peer.VisibleName(), peer.ID())

		pw := prog.New(progress.FormatNumber)
		pw.SetUpdateFrequency(200 * time.Millisecond)
		pw.Style().Visibility.TrackerOverall = false
		pw.Style().Visibility.ETA = false
		pw.Style().Visibility.Percentage = false

		tracker := prog.AppendTracker(pw, progress.FormatNumber, fmt.Sprintf("%s-%d", peer.VisibleName(), peer.ID()), 0)

		go pw.Render()

		batchSize := 100
		count := int64(0)

		builder := query.Messages(c.API()).GetHistory(peer.InputPeer()).BatchSize(batchSize)
		if opts.Time {
			builder = builder.OffsetDate(opts.To + 1)
		}
		if opts.Msg {
			builder = builder.OffsetID(opts.To + 1) // #89: retain the last msg id
		}
		iter := builder.Iter()

		// TODO(iyear): temp solution for calculating protected message id
		total, err := iter.Total(ctx)
		if err != nil {
			return err
		}
		pw.Log(color.MagentaString("[DEBUG] max message id of chat: %d", total))

		f, err := os.Create(opts.Output)
		if err != nil {
			return err
		}
		defer func(f *os.File) {
			_ = f.Close()
		}(f)

		_, err = f.WriteString(fmt.Sprintf(`{"id":%d,"messages":[`, peer.ID()))
		if err != nil {
			return err
		}
		defer func(f *os.File) {
			_, _ = f.Seek(-1, io.SeekEnd) // overwrite last comma
			_, _ = f.WriteString("]}")
		}(f)

		for iter.Next(ctx) {
			msg := iter.Value()
			if opts.Time && msg.Msg.GetDate() < opts.From {
				break
			}
			if opts.Msg && msg.Msg.GetID() < opts.From {
				break
			}

			m, ok := msg.Msg.(*tg.Message)
			if !ok || !utils.Telegram.FileExists(m) {
				continue
			}

			_, err = f.WriteString(fmt.Sprintf(`{"id":%d,"type":"message","file":"0"},`, m.ID))
			if err != nil {
				return err
			}

			count++
			tracker.SetValue(count)
		}

		if err = iter.Err(); err != nil {
			return err
		}

		tracker.MarkAsDone()
		for pw.IsRenderInProgress() {
			if pw.LengthActive() == 0 {
				pw.Stop()
			}
			time.Sleep(10 * time.Millisecond)
		}

		return nil
	})
}
