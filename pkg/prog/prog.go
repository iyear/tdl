package prog

import (
	"github.com/fatih/color"
	"github.com/iyear/tdl/pkg/utils"
	"github.com/jedib0t/go-pretty/v6/progress"
	"time"
)

func New() progress.Writer {
	pw := progress.NewWriter()
	pw.SetAutoStop(true)
	pw.SetTrackerLength(25)
	pw.SetMessageWidth(40)
	pw.SetStyle(progress.StyleDefault)
	pw.SetTrackerPosition(progress.PositionRight)
	pw.SetUpdateFrequency(time.Millisecond * 100)
	pw.Style().Colors = progress.StyleColorsExample
	pw.Style().Options.PercentFormat = "%4.1f%%"
	pw.Style().Options.ETAString = "Remaining"
	pw.Style().Visibility.TrackerOverall = true
	pw.Style().Visibility.ETA = true
	pw.Style().Visibility.ETAOverall = false
	pw.Style().Visibility.Speed = true
	pw.Style().Visibility.SpeedOverall = true
	pw.Style().Options.SpeedOverallFormatter = utils.Byte.FormatBinaryBytes
	pw.Style().Options.ErrorString = color.RedString("failed!")
	pw.Style().Options.DoneString = color.GreenString("done!")

	return pw
}
