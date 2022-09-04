package prog

import (
	"github.com/fatih/color"
	"github.com/iyear/tdl/pkg/utils"
	"github.com/jedib0t/go-pretty/v6/progress"
)

func AppendTracker(pw progress.Writer, message string, total int64) *progress.Tracker {
	units := progress.UnitsBytes
	units.Formatter = utils.Byte.FormatBinaryBytes

	tracker := progress.Tracker{
		Message: color.BlueString(message),
		Total:   total,
		Units:   units,
	}

	pw.AppendTracker(&tracker)

	return &tracker
}
