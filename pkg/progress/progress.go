package progress

import (
	"github.com/jedib0t/go-pretty/v6/progress"
	"time"
)

func New() progress.Writer {
	pw := progress.NewWriter()
	pw.SetAutoStop(true)
	pw.SetTrackerLength(25)
	pw.SetMessageWidth(24)
	pw.SetSortBy(progress.SortByPercentDsc)
	pw.SetStyle(progress.StyleDefault)
	pw.SetTrackerPosition(progress.PositionRight)
	pw.SetUpdateFrequency(time.Millisecond * 100)
	pw.Style().Colors = progress.StyleColorsExample
	pw.Style().Options.PercentFormat = "%4.1f%%"

	return pw
}
