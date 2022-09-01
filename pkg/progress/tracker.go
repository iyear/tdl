package progress

import "github.com/jedib0t/go-pretty/v6/progress"

func AppendTracker(pw progress.Writer, message string, total int64) *progress.Tracker {
	tracker := progress.Tracker{
		Message: message,
		Total:   total,
		Units:   progress.UnitsBytes,
	}

	pw.AppendTracker(&tracker)

	return &tracker
}
