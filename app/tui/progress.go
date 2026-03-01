package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
	"github.com/iyear/tdl/core/downloader"
)

const unknownName = "unknown"

// ProgressMsg updates the TUI with download progress
type ProgressMsg struct {
	ID         int64 // Unique ID for the download (using message ID or similar)
	Name       string
	State      downloader.ProgressState
	Total      int64
	IsFinished bool
	Err        error
	Cancel     context.CancelFunc // Add cancel function for early initialization
}

// ProgressStartMsg sets the initial total count for a batch of downloads
type ProgressStartMsg struct {
	Total int
}

// Ensure Model satisfies downloader.Progress interface
// Note: We need a structural adapter because Model is a value receiver in View/Update usually,
// and we need to send messages to the Program.
type TUIProgress struct {
	program *tea.Program
}

func NewTUIProgress(p *tea.Program) *TUIProgress {
	return &TUIProgress{program: p}
}

func (t *TUIProgress) OnStart(total int) {
	t.program.Send(ProgressStartMsg{Total: total})
}

func (t *TUIProgress) OnAdd(elem downloader.Elem) {
	// Send initial add message
	// We need to extract ID/Name from elem
	// elem is likely *iterElem which has .fromMsg.ID
	// But Elem interface is:
	// File() *telegram.Document
	// To() *os.File
	// ...

	// We'll use the file name as key for now or just broadcast
	name := unknownName
	if f, ok := elem.To().(interface{ Name() string }); ok {
		name = f.Name()
	}

	t.program.Send(ProgressMsg{
		Name:  name,
		Total: elem.File().Size(),
	})
}

func (t *TUIProgress) OnDownload(elem downloader.Elem, state downloader.ProgressState) {
	name := unknownName
	if f, ok := elem.To().(interface{ Name() string }); ok {
		name = f.Name()
	}

	t.program.Send(ProgressMsg{
		Name:  name,
		State: state,
		Total: elem.File().Size(),
	})
}

func (t *TUIProgress) OnDone(elem downloader.Elem, err error) {
	name := unknownName
	if f, ok := elem.To().(interface{ Name() string }); ok {
		name = f.Name()
	}

	t.program.Send(ProgressMsg{
		Name:       name,
		IsFinished: true,
		Err:        err,
	})

	if err == nil {
		// Send notification
		// We run this in a goroutine to avoid blocking
		go func() {
			_ = notify("Download Complete", fmt.Sprintf("%s has finished downloading.", name))
		}()
	}
}

// DownloadItem represents a single download in the list
type DownloadItem struct {
	Name           string
	Path           string // Full absolute path
	Total          int64
	Downloaded     int64
	LastDownloaded int64
	LastUpdate     time.Time
	SpeedBuffer    []float64
	StartTime      time.Time
	EndTime        time.Time
	Progress       progress.Model
	Finished       bool
	Err            error
	Cancel         context.CancelFunc // Used to cancel the download
}

func (d *DownloadItem) Title() string {
	return d.Name
}

func (d *DownloadItem) Description() string {
	if d.Finished {
		if d.Err != nil {
			return "❌ Failed: " + d.Err.Error()
		}
		duration := d.EndTime.Sub(d.StartTime).Round(time.Second)
		if duration < 0 {
			duration = 0
		}
		speed := float64(d.Total) / duration.Seconds()
		if duration.Seconds() == 0 {
			speed = float64(d.Total)
		}

		// Styles
		green := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
		cyan := lipgloss.NewStyle().Foreground(lipgloss.Color("6"))

		speedStr := cyan.Render(humanize.Bytes(uint64(speed)) + "/s")

		return fmt.Sprintf("%s Completed in %s (%s)", green.Render("✅"), duration, speedStr)
	}

	// Calculate Speed & ETA
	elapsed := time.Since(d.StartTime).Seconds()
	var speed float64
	var eta string

	if len(d.SpeedBuffer) > 0 {
		speed = d.SpeedBuffer[len(d.SpeedBuffer)-1]
	} else if elapsed > 0 {
		speed = float64(d.Downloaded) / elapsed // fallback to total average
	}

	if speed > 0 && d.Total > d.Downloaded {
		remainingBytes := d.Total - d.Downloaded
		remainingSeconds := float64(remainingBytes) / speed
		etaDuration := time.Duration(remainingSeconds) * time.Second
		eta = etaDuration.Round(time.Second).String()
	} else {
		eta = "∞"
	}

	// Generate Sparkline
	// Generate Sparkline
	sparkline := ""
	if len(d.SpeedBuffer) > 0 {
		bars := []rune(" ▂▃▄▅▆▇█")
		var max float64
		for _, v := range d.SpeedBuffer {
			if v > max {
				max = v
			}
		}
		var sb strings.Builder
		for _, v := range d.SpeedBuffer {
			if max == 0 || v <= 0 {
				sb.WriteRune(bars[0])
				continue
			}
			idx := int((v / max) * float64(len(bars)-1))
			if idx < 0 {
				idx = 0
			}
			if idx >= len(bars) {
				idx = len(bars) - 1
			}
			sb.WriteRune(bars[idx])
		}
		// Pad left with spaces if less than 10
		pad := 10 - len(d.SpeedBuffer)
		if pad < 0 {
			pad = 0
		}
		sparkline = strings.Repeat(" ", pad) + sb.String()
	} else {
		sparkline = "          "
	}

	// Styles
	green := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))    // Green
	cyan := lipgloss.NewStyle().Foreground(lipgloss.Color("6"))     // Cyan
	orange := lipgloss.NewStyle().Foreground(lipgloss.Color("208")) // Orange
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))    // Dim Gray
	sparkStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("135"))

	prog := d.Progress.ViewAs(d.percent())
	sparkStr := sparkStyle.Render(sparkline)
	speedStr := cyan.Render(humanize.Bytes(uint64(speed)) + "/s")
	downloadedStr := green.Render(humanize.Bytes(uint64(d.Downloaded)))
	totalStr := dim.Render("/ " + humanize.Bytes(uint64(d.Total)))

	etaStr := orange.Render("ETA: " + eta)

	return fmt.Sprintf("%s %s %s • %s • %s %s", prog, sparkStr, speedStr, etaStr, downloadedStr, totalStr)
}

func (d *DownloadItem) FilterValue() string {
	return d.Name
}

func (d *DownloadItem) percent() float64 {
	if d.Total <= 0 {
		return 0
	}
	return float64(d.Downloaded) / float64(d.Total)
}
