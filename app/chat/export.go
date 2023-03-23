package chat

import (
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/go-faster/jx"
	"github.com/gotd/contrib/middleware/ratelimit"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/tg"
	"github.com/iyear/tdl/app/internal/tgc"
	"github.com/iyear/tdl/pkg/prog"
	"github.com/iyear/tdl/pkg/storage"
	"github.com/iyear/tdl/pkg/tmedia"
	"github.com/iyear/tdl/pkg/utils"
	"github.com/jedib0t/go-pretty/v6/progress"
	"go.uber.org/multierr"
	"golang.org/x/time/rate"
	"os"
	"regexp"
	"time"
)

const (
	rateInterval = 550 * time.Millisecond
	rateBucket   = 2
)

type ExportOptions struct {
	Type   string
	Chat   string
	Input  []int
	Output string
	Filter map[string]string
}

const (
	ExportTypeTime string = "time"
	ExportTypeID   string = "id"
	ExportTypeLast string = "last"
)

var Filters = []string{FilterFile, FilterContent}

const (
	FilterFile    = "file"
	FilterContent = "content"
)

func Export(ctx context.Context, opts *ExportOptions) error {
	c, kvd, err := tgc.NoLogin(ctx, ratelimit.New(rate.Every(rateInterval), rateBucket))
	if err != nil {
		return err
	}

	return tgc.RunWithAuth(ctx, c, func(ctx context.Context) (rerr error) {
		manager := peers.Options{Storage: storage.NewPeers(kvd)}.Build(c.API())
		peer, err := utils.Telegram.GetInputPeer(ctx, manager, opts.Chat)
		if err != nil {
			return err
		}

		color.Yellow("WARN: Export only generates minimal JSON for tdl download, not for backup.")
		color.Cyan("Occasional suspensions are due to Telegram rate limitations, please wait a moment.")
		fmt.Println()

		color.Blue("Type: %s | Input: %v", opts.Type, opts.Input)

		pw := prog.New(progress.FormatNumber)
		pw.SetUpdateFrequency(200 * time.Millisecond)
		pw.Style().Visibility.TrackerOverall = false
		pw.Style().Visibility.ETA = false
		pw.Style().Visibility.Percentage = false

		tracker := prog.AppendTracker(pw, progress.FormatNumber, fmt.Sprintf("%s-%d", peer.VisibleName(), peer.ID()), 0)

		go pw.Render()

		batchSize := 100
		builder := query.Messages(c.API()).GetHistory(peer.InputPeer()).BatchSize(batchSize)
		switch opts.Type {
		case ExportTypeTime:
			builder = builder.OffsetDate(opts.Input[1] + 1)
		case ExportTypeID:
			builder = builder.OffsetID(opts.Input[1] + 1) // #89: retain the last msg id
		case ExportTypeLast:
			//builder = builder.OffsetID()
		}
		iter := builder.Iter()

		f, err := os.Create(opts.Output)
		if err != nil {
			return err
		}
		defer multierr.AppendInvoke(&rerr, multierr.Close(f))

		enc := jx.NewStreamingEncoder(f, 512)
		defer multierr.AppendInvoke(&rerr, multierr.Close(enc))

		enc.ObjStart()
		defer enc.ObjEnd()
		enc.Field("id", func(e *jx.Encoder) { e.Int64(peer.ID()) })

		enc.FieldStart("messages")
		enc.ArrStart()
		defer enc.ArrEnd()

		count := int64(0)
		re := regexpGroup(opts.Filter)
		color.Blue("Filters: %v", re)

	loop:
		for iter.Next(ctx) {
			msg := iter.Value()
			switch opts.Type {
			case ExportTypeTime:
				if msg.Msg.GetDate() < opts.Input[0] {
					break loop
				}
			case ExportTypeID:
				if msg.Msg.GetID() < opts.Input[0] {
					break loop
				}
			case ExportTypeLast:
				if count >= int64(opts.Input[0]) {
					break loop
				}
			}

			m, ok := msg.Msg.(*tg.Message)
			// filter by message content
			if !ok || !re[FilterContent].MatchString(m.Message) {
				continue
			}

			// filter by file name
			if md, ok := tmedia.GetMedia(m); !ok || !re[FilterFile].MatchString(md.Name) {
				continue
			}

			enc.Obj(func(e *jx.Encoder) {
				e.Field("id", func(e *jx.Encoder) { e.Int(m.ID) })
				e.Field("type", func(e *jx.Encoder) { e.Str("message") })
				// just a placeholder
				e.Field("file", func(e *jx.Encoder) { e.Str("0") })
			})

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

// regexpGroup returns a map of regexp.Regexp. If the value is not a valid regexp, it will be replaced with a regexp that matches everything.
func regexpGroup(m map[string]string) map[string]*regexp.Regexp {
	r := make(map[string]*regexp.Regexp)
	for k, v := range m {
		re, err := regexp.Compile(v)
		if err != nil {
			r[k] = regexp.MustCompile(".*")
			continue
		}
		r[k] = re
	}
	return r
}
