package tui

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/iyear/tdl/core/forwarder"
)

// ForwardProgressMsg updates the TUI with forwarding progress
type ForwardProgressMsg struct {
	ID         int64
	Name       string
	State      forwarder.ProgressState
	IsFinished bool
	Err        error
}

// Ensure TUIForwardProgress satisfies forwarder.Progress
type TUIForwardProgress struct {
	program    *tea.Program
	lastUpdate time.Time
	mu         sync.Mutex
}

func NewTUIForwardProgress(p *tea.Program) *TUIForwardProgress {
	return &TUIForwardProgress{
		program:    p,
		lastUpdate: time.Now(),
	}
}

func (t *TUIForwardProgress) OnAdd(elem forwarder.Elem) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Only send Add updates occasionally to prevent Native Batch UI lag
	if time.Since(t.lastUpdate) > 100*time.Millisecond {
		name := fmt.Sprintf("[%d] %d", elem.From().ID(), elem.Msg().ID)
		t.program.Send(ForwardProgressMsg{
			ID:   int64(elem.Msg().ID),
			Name: name,
		})
		t.lastUpdate = time.Now()
	}
}

func (t *TUIForwardProgress) OnClone(elem forwarder.Elem, state forwarder.ProgressState) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if time.Since(t.lastUpdate) > 100*time.Millisecond {
		name := fmt.Sprintf("[%d] %d", elem.From().ID(), elem.Msg().ID)
		t.program.Send(ForwardProgressMsg{
			ID:    int64(elem.Msg().ID),
			Name:  name,
			State: state,
		})
		t.lastUpdate = time.Now()
	}
}

func (t *TUIForwardProgress) OnDone(elem forwarder.Elem, err error) {
	// Always send Done messages to clear UI status
	name := fmt.Sprintf("[%d] %d", elem.From().ID(), elem.Msg().ID)

	t.program.Send(ForwardProgressMsg{
		ID:         int64(elem.Msg().ID),
		Name:       name,
		IsFinished: true,
		Err:        err,
	})

	if err == nil {
		go func() {
			_ = notify("Forward Complete", fmt.Sprintf("Message %d forwarded.", elem.Msg().ID))
		}()
	}
}

// ForwardItem represents a single forwarded message cloning
type ForwardItem struct {
	ID             int64
	Name           string
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
	Cancel         context.CancelFunc
}

func (f *ForwardItem) Title() string {
	return f.Name
}

func (f *ForwardItem) Description() string {
	// Re-use same formatting logic from DownloadItem for Sparklines and ETA
	di := &DownloadItem{
		Name:           f.Name,
		Total:          f.Total,
		Downloaded:     f.Downloaded,
		LastDownloaded: f.LastDownloaded,
		LastUpdate:     f.LastUpdate,
		SpeedBuffer:    f.SpeedBuffer,
		StartTime:      f.StartTime,
		EndTime:        f.EndTime,
		Progress:       f.Progress,
		Finished:       f.Finished,
		Err:            f.Err,
	}
	return di.Description()
}

func (f *ForwardItem) FilterValue() string {
	return f.Name
}
