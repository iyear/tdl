package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Download Options Intercept
	if m.state == stateDownloadOptions {
		return m.updateDownloadOptions(msg)
	}

	// Global Config Editor Intercept
	if m.state == stateConfig {
		return m.updateConfig(msg)
	}

	// Batch File Picker Intercept
	if m.state == stateBatch {
		var cmd tea.Cmd
		m.FilePicker, cmd = m.FilePicker.Update(msg)

		if didSelect, path := m.FilePicker.DidSelectFile(msg); didSelect {
			// Direct to Options Form
			m.BatchPath = path
			m.state = stateDownloadOptions
			m.DLForm.UrlOrPath = m.BatchPath
			m.DLForm.IsBatch = true
			m.DLForm.ActiveIndex = 0
			m.DLForm.Dir.Focus()
			return m, nil
		}

		// Handle Esc to exit
		if msg, ok := msg.(tea.KeyMsg); ok && msg.String() == "esc" {
			m.state = stateDashboard
		}

		return m, cmd
	}

	// Batch Confirm Intercept
	if m.state == stateBatchConfirm {
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "d", "enter":
				// Batch Confirm -> Options Form
				m.state = stateDownloadOptions
				m.DLForm.UrlOrPath = m.BatchPath
				m.DLForm.IsBatch = true
				m.DLForm.ActiveIndex = 0
				m.DLForm.Dir.Focus()
				return m, nil
			case "f":
				m.PickingDest = true
				m.ForwardSource = []string{m.BatchPath}
				m.state = stateDashboard
				m.ActiveTab = 1 // Browser
				m.Pane = 0      // Dialogs
				m.StatusMessage = "Select destination chat for JSON batch..."
				// Trigger dialog fetch if needed
				if len(m.Dialogs.Items()) == 0 {
					m.LoadingDialogs = true
					return m, tea.Batch(m.GetDialogs(nil, 0, 0), m.spinner.Tick)
				}
				return m, nil
			case "esc", "q":
				m.state = stateBatch
				m.BatchPath = ""
				return m, nil
			}
		}
		return m, nil
	}

	// Export Prompt Intercept
	if m.state == stateExportPrompt {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				filename := m.ExportInput.Value()
				if filename == "" {
					filename = fmt.Sprintf("%d.json", m.ExportTarget.PeerID)
				}
				// Ensure .json extension
				if len(filename) < 5 || filename[len(filename)-5:] != ".json" {
					filename += ".json"
				}

				m.state = stateDashboard
				m.ActiveTab = 1 // Return to browser
				m.LoadingExport = true
				m.StatusMessage = "Exporting to " + filename + "..."
				return m, tea.Batch(m.startExport(m.ExportTarget, filename), m.spinner.Tick)

			case "esc":
				m.state = stateDashboard
				m.ActiveTab = 1
				m.StatusMessage = "Export canceled"
				return m, nil
			}
		}
		var cmd tea.Cmd
		m.ExportInput, cmd = m.ExportInput.Update(msg)
		return m, cmd
	}

	switch msg := msg.(type) {
	case loginMsg:
		if msg.Err != nil {
			m.Connected = false
		} else {
			m.Connected = true
			m.User = msg.User
		}
	case ProgressMsg:
		item, exists := m.Downloads[msg.Name]
		if !exists {
			// New download
			prog := progress.New(progress.WithDefaultGradient())
			item = &DownloadItem{
				Name:      msg.Name,
				Path:      msg.Name, // Assuming msg.Name is the file path
				Total:     msg.Total,
				StartTime: time.Now(),
				Progress:  prog,
			}
			// Use Base name for display if it looks like a path
			if filepath.IsAbs(msg.Name) || len(filepath.Dir(msg.Name)) > 1 {
				// We keep Name as full path for map key?
				// Actually typically msg.Name from TUIProgress is what we get.
				// Let's rely on Title() doing Base() if we want, or here.
				// For now simple.
			}

			m.Downloads[msg.Name] = item
			m.DownloadList.InsertItem(0, item)
		} else {
			// Update existing
			item.Downloaded = msg.State.Downloaded
			if msg.Total > 0 {
				item.Total = msg.Total
			}

			if msg.IsFinished {
				item.Finished = true
				item.Err = msg.Err
			}
		}

	case dialogsMsg:
		m.LoadingDialogs = false

		// Update Offsets
		m.NextOffsetPeer = msg.NextPeer
		m.NextOffsetDate = msg.NextDate
		m.NextOffsetID = msg.NextID

		if msg.Err != nil {
			// Handle error
		} else {
			items := make([]list.Item, len(msg.Dialogs))
			for i, d := range msg.Dialogs {
				items[i] = d
			}

			if m.IsPaginating {
				// Append
				// Create new slice with old + new
				// list.Model doesn't have Append? It has InsertItem.
				// Or SetItems with combined list.
				oldItems := m.Dialogs.Items()
				newItems := append(oldItems, items...)
				m.Dialogs.SetItems(newItems)
				m.IsPaginating = false
			} else {
				m.Dialogs.SetItems(items)
			}
		}

	case historyMsg:
		m.LoadingHistory = false
		if msg.Err != nil {
			// Handle error
		} else {
			items := make([]list.Item, len(msg.Messages))
			for i, m := range msg.Messages {
				items[i] = m
			}
			m.Messages.SetItems(items)
		}

	case ExportProgressMsg:
		m.StatusMessage = fmt.Sprintf("Exporting... %d messages processed", int64(msg))
		return m, nil

	case ExportMsg:
		m.LoadingExport = false
		if msg.Err != nil {
			m.StatusMessage = fmt.Sprintf("Export Failed: %v", msg.Err)
		} else {
			m.StatusMessage = fmt.Sprintf("Exported to %s", msg.Path)
		}

		if msg.Err != nil {
			// handle global or item specific error
		}

		// Fallthrough only if we have logic for item updates here
		// But in this case we seem to have mixed logic from ProgressMsg.
		// The original ProgressMsg block handles 'item'.

		return m, nil

		// Update progress bar model
		// Calculate percentage
		// Update progress bar model
		// Calculate percentage
		// var pct float64
		// if item.Total > 0 {
		// 	pct = float64(item.Downloaded) / float64(item.Total)
		// }
		// We don't really have a cmd from progress update usually unless it animates
		// But here we just set percentage for view
		// Actually bubbles/progress needs an update msg for animation, but we can just View() it with strict percentage if we want
		// or use SetPercent

		// For now simple reliable approach:
		// We are not using the bubble's internal ticking for smooth animation to keep it simple first

		return m, nil
	case AccountsMsg:
		if msg.Err == nil {
			m.Accounts = msg.Accounts
		}

	case AccountSwitchedMsg:
		if msg.Err != nil {
			m.StatusMessage = fmt.Sprintf("Switch Failed: %v", msg.Err)
		} else {
			m.Namespace = msg.Namespace
			m.storage = msg.Storage
			m.StatusMessage = fmt.Sprintf("Switched to %s", msg.Namespace)

			// Reset State
			m.User = nil
			m.Connected = false
			m.Dialogs.SetItems(nil)
			m.Messages.SetItems(nil)

			// Re-login
			return m, m.startClient
		}

	case tea.KeyMsg:
		// 1. Priority Globals
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "?":
			m.ShowHelp = !m.ShowHelp
			return m, nil
		case "esc":
			if m.ShowHelp {
				m.ShowHelp = false
				return m, nil
			}
			// History Pop
			if len(m.TabHistory) > 0 {
				last := m.TabHistory[len(m.TabHistory)-1]
				m.TabHistory = m.TabHistory[:len(m.TabHistory)-1]
				m.ActiveTab = last
				// Restore State
				switch last {
				case 0:
					m.state = stateDashboard
				case 1:
					m.state = stateDashboard // Browser shares state usually or we can verify
				case 2:
					m.state = stateDownloads
				}
				return m, nil
			}
			// Fallthrough to global quit checking if history empty?
			// Or just do nothing.

		}

		// 2. Global Navigation (Safe Keys)
		// j is excluded here to allow list navigation in Browser
		switch msg.String() {
		case "d":
			if m.ActiveTab != 0 {
				m.TabHistory = append(m.TabHistory, m.ActiveTab)
				m.state = stateDashboard
				m.ActiveTab = 0
			}
			return m, nil
		case "b":
			if m.ActiveTab != 1 {
				m.TabHistory = append(m.TabHistory, m.ActiveTab)
				m.ActiveTab = 1
				if len(m.Dialogs.Items()) == 0 {
					m.LoadingDialogs = true
					return m, tea.Batch(m.GetDialogs(nil, 0, 0), m.spinner.Tick)
				}
			}
			return m, nil
		case "L":
			if m.ActiveTab == 1 && !m.PickingDest {
				m.IsPaginating = true
				m.LoadingDialogs = true
				return m, tea.Batch(
					m.GetDialogs(m.NextOffsetPeer, m.NextOffsetDate, m.NextOffsetID),
					m.spinner.Tick,
				)
			}
			return m, nil
		case "l":
			if m.ActiveTab != 2 {
				m.TabHistory = append(m.TabHistory, m.ActiveTab)
				m.state = stateDownloads
				m.ActiveTab = 2
			}
			return m, nil
		case "c":
			m.state = stateConfig
			m.InitConfigInputs()
			m.ConfigFocusIndex = 0
			return m, nil
		case "i":
			m.input.Focus()
			return m, textinput.Blink
		case "a":
			if len(m.Accounts) > 1 {
				idx := -1
				for i, acc := range m.Accounts {
					if acc == m.Namespace {
						idx = i
						break
					}
				}
				nextIdx := (idx + 1) % len(m.Accounts)
				return m, m.SwitchAccount(m.Accounts[nextIdx])
			}
			return m, nil
		case "tab":
			if m.ActiveTab == 1 {
				m.Pane = 1 - m.Pane
				return m, nil
			}
		}

		// If input is focused, pass messages to it
		if m.input.Focused() {
			switch msg.String() {
			case "enter":
				val := m.input.Value()
				m.input.Reset()
				m.input.Blur()

				if m.Searching {
					m.Searching = false
					m.LoadingDialogs = true
					m.Pane = 0
					m.StatusMessage = "Searching: " + val
					return m, tea.Batch(m.SearchPeers(val), m.spinner.Tick)
				}

				m.state = stateDownloadOptions
				m.DLForm.UrlOrPath = val
				m.DLForm.IsBatch = false
				m.DLForm.ActiveIndex = 0
				m.DLForm.Dir.Focus()
				return m, nil
			case "esc":
				m.Searching = false
				m.input.Blur()
				return m, nil
			}
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}

		// List handling if in Browser
		if m.ActiveTab == 1 {
			var cmd tea.Cmd
			if m.Pane == 0 {
				m.Dialogs, cmd = m.Dialogs.Update(msg)
				// Handle Enter
				if msg.String() == "enter" {
					if m.PickingDest {
						// Execute Forward
						if dlg, ok := m.Dialogs.SelectedItem().(DialogItem); ok {
							dest := strconv.FormatInt(dlg.PeerID, 10) // Use ID as dest
							sources := m.ForwardSource

							// Reset state
							m.PickingDest = false
							m.ForwardSource = nil
							m.StatusMessage = fmt.Sprintf("Forwarding to %s...", dlg.Title)

							return m, m.startForward(dest, sources)
						}
					}

					// Fetch history for selected dialog
					if dlg, ok := m.Dialogs.SelectedItem().(DialogItem); ok {
						m.Messages.SetItems(nil) // Clear previous
						m.LoadingHistory = true
						return m, tea.Batch(m.GetHistory(dlg.Peer), m.spinner.Tick)
					}
					m.Pane = 1
				}
				// Handle Export
				if msg.String() == "e" && !m.PickingDest {
					if dlg, ok := m.Dialogs.SelectedItem().(DialogItem); ok {
						m.state = stateExportPrompt
						m.ExportTarget = dlg
						m.ExportInput.Reset()
						m.ExportInput.SetValue(fmt.Sprintf("%d.json", dlg.PeerID))
						m.ExportInput.Focus()
						return m, textinput.Blink
					}
				}
				return m, cmd
			} else {
				m.Messages, cmd = m.Messages.Update(msg)

				// Message Selection (Space)
				if msg.String() == " " {
					if idx := m.Messages.Index(); idx >= 0 {
						if item, ok := m.Messages.SelectedItem().(MessageItem); ok {
							item.Selected = !item.Selected
							m.Messages.SetItem(idx, item)
							return m, nil
						}
					}
				}

				// Forward Init (f)
				if msg.String() == "f" {
					// ... (existing forward logic)
					// Collect selected
					var sources []string
					for _, item := range m.Messages.Items() {
						if mItem, ok := item.(MessageItem); ok && mItem.Selected {
							// Construct link
							link := fmt.Sprintf("https://t.me/c/%d/%d", mItem.ChatID, mItem.ID)
							sources = append(sources, link)
						}
					}

					if len(sources) > 0 {
						m.PickingDest = true
						m.ForwardSource = sources
						m.Pane = 0 // Switch to dialogs to pick
						m.StatusMessage = fmt.Sprintf("Select destination chat for %d messages...", len(sources))
						return m, nil
					}
					m.StatusMessage = "No messages selected. Use [Space] to select."
					return m, nil
				}

				return m, cmd
			}
		}

		switch msg.String() {
		case "j":
			m.state = stateBatch
			m.FilePicker.CurrentDirectory, _ = os.Getwd() // Reset to cwd
			return m, m.FilePicker.Init()
		case "s":
			if m.ActiveTab == 1 { // Browser
				m.Searching = true
				m.input.Placeholder = "Search Global... (Enter to submit)"
				m.input.Focus()
				return m, textinput.Blink
			}
		}
	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			// Tab Hit Testing (Approximate)
			// Header is ~2 lines, padding ~1 line. Tabs usually around line 3-5.
			if msg.Y >= 2 && msg.Y <= 6 {
				// Tabs are left aligned: Dashboard | Browser | Downloads
				// Dashboard (9+2=11), Browser (7+2=9), Downloads (9+2=11)
				// Ranges: 0-11, 11-20, 20-31
				if msg.X >= 0 && msg.X < 12 {
					if m.ActiveTab != 0 {
						m.TabHistory = append(m.TabHistory, m.ActiveTab)
						m.ActiveTab = 0
						m.state = stateDashboard
					}
					return m, nil
				} else if msg.X >= 12 && msg.X < 22 {
					if m.ActiveTab != 1 {
						m.TabHistory = append(m.TabHistory, m.ActiveTab)
						m.ActiveTab = 1
						// Trigger fetch dialogs if empty
						if len(m.Dialogs.Items()) == 0 {
							m.LoadingDialogs = true
							return m, tea.Batch(m.GetDialogs(nil, 0, 0), m.spinner.Tick)
						}
					}
					return m, nil
				} else if msg.X >= 22 && msg.X < 40 {
					if m.ActiveTab != 2 {
						m.TabHistory = append(m.TabHistory, m.ActiveTab)
						m.ActiveTab = 2
						m.state = stateDownloads
					}
					return m, nil
				}
			}
		}

		// Forward mouse to active components
		var cmd tea.Cmd
		if m.ActiveTab == 1 {
			if m.Pane == 0 {
				m.Dialogs, cmd = m.Dialogs.Update(msg)
				return m, cmd
			} else {
				m.Messages, cmd = m.Messages.Update(msg)
				return m, cmd
			}
		} else if m.ActiveTab == 2 {
			var cmd tea.Cmd
			m.DownloadList, cmd = m.DownloadList.Update(msg)

			if msg.String() == "o" {
				if item, ok := m.DownloadList.SelectedItem().(*DownloadItem); ok {
					if err := openFile(item.Path); err != nil {
						m.StatusMessage = fmt.Sprintf("Failed to open: %v", err)
					} else {
						m.StatusMessage = fmt.Sprintf("Opening %s...", filepath.Base(item.Path))
					}
				}
			}
			return m, cmd
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height

		// Calculate available space
		// Borders: 2 (Left) + 2 (Right) = 4
		// Margin: 1
		// Total Overhead: 5
		availableWidth := m.width - 5
		if availableWidth < 0 {
			availableWidth = 0
		}

		leftWidth := availableWidth / 3
		rightWidth := availableWidth - leftWidth

		// Height Overhead: Header (~3) + Tabs (~3) + Footer (~3) = ~9
		// Using -10 to be safe
		listHeight := m.height - 10
		if listHeight < 0 {
			listHeight = 0
		}

		// Resize lists
		m.Dialogs.SetSize(leftWidth, listHeight)
		m.Messages.SetSize(rightWidth, listHeight)
		m.DownloadList.SetSize(m.width-2, listHeight) // Full width - border

		// Resize FilePicker
		m.FilePicker.Height = listHeight

	}

	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m *Model) updateDownloadOptions(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.state = stateDownloads
			if m.ActiveTab == 1 {
				m.state = stateDashboard
			} // Return to browser if that's where we came from?
			// Actually let's just return to where we were.
			// Ideally we track previous state, but usually it's Dashboard (Browser) or Downloads tab.
			// Simple behavior: Go to Downloads tab.
			m.ActiveTab = 2
			m.state = stateDownloads
			return m, nil

		case "up", "shift+tab":
			m.DLForm.ActiveIndex--
			if m.DLForm.ActiveIndex < 0 {
				m.DLForm.ActiveIndex = 14
			}
			return m, m.updateFocus()

		case "down", "tab":
			m.DLForm.ActiveIndex++
			if m.DLForm.ActiveIndex > 14 {
				m.DLForm.ActiveIndex = 0
			}
			return m, m.updateFocus()

		case "enter":
			// Action based on index
			switch m.DLForm.ActiveIndex {
			case 2:
				m.DLForm.Group = !m.DLForm.Group
			case 3:
				m.DLForm.SkipSame = !m.DLForm.SkipSame
			case 4:
				m.DLForm.Takeout = !m.DLForm.Takeout
			case 5:
				m.DLForm.Desc = !m.DLForm.Desc
			// 6-10 are text inputs, enter moves focus
			case 11:
				m.DLForm.Continue = !m.DLForm.Continue
			case 12:
				m.DLForm.Debug = !m.DLForm.Debug
			case 13: // Start
				m.state = stateDownloads
				m.ActiveTab = 2

				// Parse Advanced Options
				// We don't have separate fields in startDownload signature,
				// so we need to pass a full struct or modify startDownload to take DLForm?
				// Better: startDownload reads from m.DLForm!

				if m.DLForm.IsBatch {
					return m, m.startBatchDownload(m.DLForm.UrlOrPath)
				}
				return m, m.startDownload(m.DLForm.UrlOrPath)
			case 14: // Cancel
				m.state = stateDownloads
				return m, nil
			}
			// If on text inputs, Enter might move next?
			if m.DLForm.ActiveIndex <= 1 || (m.DLForm.ActiveIndex >= 6 && m.DLForm.ActiveIndex <= 10) {
				m.DLForm.ActiveIndex++
				return m, m.updateFocus()
			}
			return m, nil
		}
	}

	// Handle Text Input Updates
	var cmd tea.Cmd
	switch m.DLForm.ActiveIndex {
	case 0:
		m.DLForm.Dir, cmd = m.DLForm.Dir.Update(msg)
	case 1:
		m.DLForm.Template, cmd = m.DLForm.Template.Update(msg)
	case 6:
		m.DLForm.Threads, cmd = m.DLForm.Threads.Update(msg)
	case 7:
		m.DLForm.Limit, cmd = m.DLForm.Limit.Update(msg)
	case 8:
		m.DLForm.Pool, cmd = m.DLForm.Pool.Update(msg)
	case 9:
		m.DLForm.Delay, cmd = m.DLForm.Delay.Update(msg)
	case 10:
		m.DLForm.Reconnect, cmd = m.DLForm.Reconnect.Update(msg)
	}

	if cmd != nil {
		return m, cmd
	}

	return m, nil
}

func (m *Model) updateFocus() tea.Cmd {
	// Blur all
	m.DLForm.Dir.Blur()
	m.DLForm.Template.Blur()
	m.DLForm.Threads.Blur()
	m.DLForm.Limit.Blur()
	m.DLForm.Pool.Blur()
	m.DLForm.Delay.Blur()
	m.DLForm.Reconnect.Blur()

	var cmd tea.Cmd
	switch m.DLForm.ActiveIndex {
	case 0:
		cmd = m.DLForm.Dir.Focus()
	case 1:
		cmd = m.DLForm.Template.Focus()
	case 6:
		cmd = m.DLForm.Threads.Focus()
	case 7:
		cmd = m.DLForm.Limit.Focus()
	case 8:
		cmd = m.DLForm.Pool.Focus()
	case 9:
		cmd = m.DLForm.Delay.Focus()
	case 10:
		cmd = m.DLForm.Reconnect.Focus()
	}
	return cmd
}
