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
	"github.com/iyear/tdl/pkg/utils"
	"github.com/jedib0t/go-pretty/v6/progress"
	"golang.org/x/time/rate"
	"os"
	"time"
)

const (
	layout       = "2006-01-02 15:04:05"
	rateInterval = 475 * time.Millisecond
	rateBucket   = 2
)

func Export(ctx context.Context, chat string, from, to int, media bool, output string) error {
	c, _, err := tgc.NoLogin(ratelimit.New(rate.Every(rateInterval), rateBucket))
	if err != nil {
		return err
	}

	return tgc.RunWithAuth(ctx, c, func(ctx context.Context) error {
		manager := peers.Options{}.Build(c.API())
		peer, err := utils.Telegram.GetInputPeer(ctx, manager, chat)
		if err != nil {
			return err
		}

		color.Blue("Indexing... %s/%d: %s ~ %s", peer.VisibleName(), peer.ID(), time.Unix(int64(from), 0).Format(layout), time.Unix(int64(to), 0).Format(layout))
		color.Cyan("Fetch speed is determined by Telegram server")

		pw := prog.New(progress.FormatNumber)
		pw.SetUpdateFrequency(200 * time.Millisecond)
		pw.Style().Visibility.TrackerOverall = false
		pw.Style().Visibility.ETA = false
		pw.Style().Visibility.Percentage = false

		tracker := prog.AppendTracker(pw, progress.FormatNumber, fmt.Sprintf("%s-%d", peer.VisibleName(), peer.ID()), 0)

		go pw.Render()

		batchSize := 100
		count := int64(0)
		iter := query.Messages(c.API()).GetHistory(peer.InputPeer()).
			OffsetDate(to).BatchSize(batchSize).Iter()

		f, err := os.Create(output)
		if err != nil {
			return err
		}
		_ = f

		for iter.Next(ctx) {
			msg := iter.Value()
			if msg.Msg.GetDate() < from {
				break
			}

			m, ok := msg.Msg.(*tg.Message)
			if !ok {
				continue
			}

			_, ok = m.GetMedia()
			if media && !ok {
				continue
			}

			count++
			tracker.SetValue(count)
		}

		return nil
	})
}
