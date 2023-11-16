package prog

import (
	"time"

	"github.com/fatih/color"
	"github.com/go-faster/errors"
	"github.com/jedib0t/go-pretty/v6/progress"
	"github.com/jedib0t/go-pretty/v6/text"
	tsize "github.com/kopoli/go-terminal-size"
)

func New(formatter progress.UnitsFormatter) (progress.Writer, error) {
	pw := progress.NewWriter()
	pw.SetAutoStop(false)

	size, err := tsize.GetSize()
	if err != nil {
		return nil, errors.Wrap(err, "get terminal size")
	}
	width := size.Width
	if width > 100 {
		width = 100
	}
	pw.SetTrackerLength(width / 3)
	pw.SetMessageWidth(width * 2 / 3)
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

	return pw, nil
}

func Wait(pw progress.Writer) {
	for pw.IsRenderInProgress() {
		if pw.LengthActive() == 0 {
			pw.Stop()
		}
		time.Sleep(10 * time.Millisecond)
	}
}
