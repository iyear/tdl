package prog

import (
	"context"
	"time"

	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/progress"
	"github.com/jedib0t/go-pretty/v6/text"
	tsize "github.com/kopoli/go-terminal-size"
)

func New(formatter progress.UnitsFormatter) progress.Writer {
	pw := progress.NewWriter()
	pw.SetAutoStop(false)

	width := 100
	if size, err := tsize.GetSize(); err == nil {
		width = size.Width
	}
	width -= 50 // tail length
	pw.SetTrackerLength(width / 5)
	pw.SetMessageWidth(width * 3 / 5)
	pw.SetStyle(progress.StyleDefault)
	pw.SetTrackerPosition(progress.PositionRight)
	pw.SetUpdateFrequency(time.Millisecond * 100)
	pw.Style().Colors = progress.StyleColorsExample
	pw.Style().Colors.Message = text.Colors{text.FgBlue}
	pw.Style().Options.PercentFormat = "%4.1f%%"
	pw.Style().Visibility.TrackerOverall = true
	pw.Style().Visibility.ETA = true
	pw.Style().Visibility.ETAOverall = false
	pw.Style().Visibility.Speed = true
	pw.Style().Visibility.SpeedOverall = true
	pw.Style().Visibility.Pinned = true
	pw.Style().Options.TimeInProgressPrecision = time.Millisecond
	pw.Style().Options.SpeedOverallFormatter = formatter
	pw.Style().Options.ErrorString = color.RedString("failed!")
	pw.Style().Options.DoneString = color.GreenString("done!")

	return pw
}

func Wait(ctx context.Context, pw progress.Writer) {
	for pw.IsRenderInProgress() {
		select {
		case <-ctx.Done():
			return
		default:
			if pw.LengthActive() == 0 {
				pw.Stop()
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
}
