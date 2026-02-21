package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

	// Login Flow Intercept
	if m.state == stateLogin || m.state == stateLoginPhone || m.state == stateLoginCode || m.state == stateLoginPassword {
		return m.updateLogin(msg)
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

	// Dir Picker Intercept
	if m.state == stateDirPicker {
		// Try to handle 's' to select the current directory
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "s":
				m.DLForm.Dir.SetValue(m.FilePicker.CurrentDirectory)
				m.state = stateDownloadOptions
				return m, nil
			case "esc":
				m.state = stateDownloadOptions
				return m, nil
			}
		}

		var cmd tea.Cmd
		m.FilePicker, cmd = m.FilePicker.Update(msg)
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

	if m.state == stateAccounts {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				if item, ok := m.AccountsList.SelectedItem().(AccountItem); ok {
					if !item.IsActive {
						m.StatusMessage = fmt.Sprintf("Switching to %s...", item.Name)
						m.state = stateDashboard
						return m, m.SwitchAccount(item.Name)
					}
				}
				return m, nil
			case "n":
				m.state = stateLoginPhone
				m.AuthPhone.Reset()
				m.AuthPhone.Focus()
				m.StatusMessage = "Adding new account via login flow."
				return m, textinput.Blink
			case "esc":
				m.state = stateDashboard
				return m, nil
			}
		}
		var cmd tea.Cmd
		m.AccountsList, cmd = m.AccountsList.Update(msg)
		return m, cmd
	}

	switch msg := msg.(type) {
	case loginMsg:
		if msg.Err != nil {
			m.Connected = false
		} else {
			m.Connected = true
			m.User = msg.User
			m.state = stateDashboard // Auto return
		}

	case sysTickMsg:
		c, mem := fetchSysMetrics()
		m.sysCpu = c
		m.sysMem = mem
		return m, sysTick()

	case authMsg:
		if msg.Err != nil {
			m.StatusMessage = fmt.Sprintf("Login Error: %v", msg.Err)
			m.state = stateLoginPhone // Reset to start on error (or we could keep state and just show error)
		} else {
			m.state = msg.State
			if msg.Hash != "" {
				m.AuthCodeHash = msg.Hash
			}
			m.StatusMessage = "" // Clear message on success
			if m.state == stateDashboard {
				m.StatusMessage = "Login Successful!"
				// Trigger a fetch to get user info since we are now logged in
				return m, m.startClient
			}
		}
	case ProgressMsg:
		item, exists := m.Downloads[msg.Name]
		if !exists {
			// New download
			prog := progress.New(progress.WithDefaultGradient())
			item = &DownloadItem{
				Name:           msg.Name,
				Path:           msg.Name, // Assuming msg.Name is the file path
				Total:          msg.Total,
				StartTime:      time.Now(),
				LastUpdate:     time.Now(),
				LastDownloaded: 0,
				Progress:       prog,
			}
			// Use Base name for display if it looks like a path
			if filepath.IsAbs(msg.Name) || len(filepath.Dir(msg.Name)) > 1 {
				// We keep Name as full path for map key?
				// Actually typically msg.Name from TUIProgress is what we get.
				// Let's rely on Title() doing Base() if we want, or here.
				// For now simple.
			}

			m.Downloads[msg.Name] = item
			// Append to the list instead of prepending, to prevent scrolling issues during batch inserts
			m.DownloadList.InsertItem(len(m.DownloadList.Items()), item)

			// Increment Batch Counters
			if m.DLForm.IsBatch {
				m.BatchTotal++
			}
		} else {
			// Update existing
			if msg.Cancel != nil {
				item.Cancel = msg.Cancel
			}

			// Live Speed Calculation
			now := time.Now()
			elapsedSinceLast := now.Sub(item.LastUpdate).Seconds()
			if elapsedSinceLast >= 1.0 {
				bytesSince := msg.State.Downloaded - item.LastDownloaded
				speed := float64(bytesSince) / elapsedSinceLast

				item.SpeedBuffer = append(item.SpeedBuffer, speed)
				if len(item.SpeedBuffer) > 10 { // keep last 10 ticks for sparklines
					item.SpeedBuffer = item.SpeedBuffer[1:]
				}

				item.LastDownloaded = msg.State.Downloaded
				item.LastUpdate = now
			}

			item.Downloaded = msg.State.Downloaded
			if msg.Total > 0 {
				item.Total = msg.Total
			}

			if msg.IsFinished {
				item.Finished = true
				item.EndTime = time.Now()
				item.Err = msg.Err
				if item.Err == nil {
					if strings.HasSuffix(item.Name, ".tmp") {
						item.Name = strings.TrimSuffix(item.Name, ".tmp")
					}
					if m.DLForm.IsBatch {
						m.BatchCompleted++
					}
				}
			}
		}

	case ForwardProgressMsg:
		var item *ForwardItem
		var exists bool

		if item, exists = m.Forwards[msg.ID]; !exists {
			// New forward clone task
			prog := progress.New(progress.WithDefaultGradient())
			item = &ForwardItem{
				ID:             msg.ID,
				Name:           msg.Name,
				Total:          msg.State.Total,
				StartTime:      time.Now(),
				LastUpdate:     time.Now(),
				LastDownloaded: 0,
				Progress:       prog,
			}
			m.Forwards[msg.ID] = item
			m.ForwardList.InsertItem(len(m.ForwardList.Items()), item)
		} else {
			// Live Speed Calculation
			now := time.Now()
			elapsedSinceLast := now.Sub(item.LastUpdate).Seconds()
			if elapsedSinceLast >= 1.0 {
				bytesSince := msg.State.Done - item.LastDownloaded
				speed := float64(bytesSince) / elapsedSinceLast

				item.SpeedBuffer = append(item.SpeedBuffer, speed)
				if len(item.SpeedBuffer) > 10 {
					item.SpeedBuffer = item.SpeedBuffer[1:]
				}

				item.LastDownloaded = msg.State.Done
				item.LastUpdate = now
			}

			item.Downloaded = msg.State.Done
			if msg.State.Total > 0 {
				item.Total = msg.State.Total
			}

			if msg.IsFinished {
				item.Finished = true
				item.EndTime = time.Now()
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

		return m, nil
	case AccountsMsg:
		if msg.Err == nil {
			m.Accounts = msg.Accounts

			var items []list.Item
			for _, acc := range m.Accounts {
				items = append(items, AccountItem{
					Name:     acc,
					IsActive: acc == m.Namespace,
				})
			}
			m.AccountsList.SetItems(items)
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
		case "w":
			if m.ActiveTab != 3 {
				m.TabHistory = append(m.TabHistory, m.ActiveTab)
				m.ActiveTab = 3
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
		case "r":
			if !m.Connected {
				m.state = stateLoginPhone
				m.AuthPhone.Reset()
				m.AuthPhone.Focus()
				return m, textinput.Blink
			}
			return m, nil
		case "a", "A":
			m.state = stateAccounts
			// Refresh list active status dynamically
			var items []list.Item
			for _, acc := range m.Accounts {
				items = append(items, AccountItem{
					Name:     acc,
					IsActive: acc == m.Namespace,
				})
			}
			m.AccountsList.SetItems(items)
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

		// List handling if in Browser or Downloads
		// Tab-specific global overrides that bypass list navigation
		switch msg.String() {
		case "j":
			if m.ActiveTab == 0 || m.ActiveTab == 2 { // Dashboard or Downloads
				m.state = stateBatch
				m.FilePicker.CurrentDirectory, _ = os.Getwd()
				return m, m.FilePicker.Init()
			}
		case "s":
			if m.ActiveTab == 1 { // Browser
				m.Searching = true
				m.input.Placeholder = "Search Global... (Enter to submit)"
				m.input.Focus()
				return m, textinput.Blink
			}
		}

		switch m.ActiveTab {
		case 1:
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

							m.TabHistory = append(m.TabHistory, m.ActiveTab)
							m.ActiveTab = 3 // Jump to Forwarding Tab automatically

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
		case 2:
			var cmd tea.Cmd
			m.DownloadList, cmd = m.DownloadList.Update(msg)

			switch msg.String() {
			case "o", "enter":
				if item, ok := m.DownloadList.SelectedItem().(*DownloadItem); ok {
					if item.Finished {
						if err := openFile(item.Path); err != nil {
							m.StatusMessage = fmt.Sprintf("Failed to open: %v", err)
						} else {
							m.StatusMessage = fmt.Sprintf("Opening %s...", filepath.Base(item.Path))
						}
					} else {
						m.StatusMessage = "Download not finished yet"
					}
				}
			case "x", "delete":
				if item, ok := m.DownloadList.SelectedItem().(*DownloadItem); ok {
					if !item.Finished && item.Cancel != nil {
						item.Cancel()
						m.StatusMessage = fmt.Sprintf("Cancelled %s", item.Name)
					} else if item.Finished {
						m.StatusMessage = fmt.Sprintf("Removed %s from list", item.Name)
					}

					// Remove from map and list
					delete(m.Downloads, item.Name)
					idx := m.DownloadList.Index()
					if idx >= 0 {
						m.DownloadList.RemoveItem(idx)
					}
				}
			}
			return m, cmd
		case 3:
			var cmd tea.Cmd
			m.ForwardList, cmd = m.ForwardList.Update(msg)

			switch msg.String() {
			case "x", "delete":
				if item, ok := m.ForwardList.SelectedItem().(*ForwardItem); ok {
					if !item.Finished && item.Cancel != nil {
						item.Cancel()
						m.StatusMessage = fmt.Sprintf("Cancelled forward cloning for %s", item.Name)
					} else if item.Finished {
						m.StatusMessage = fmt.Sprintf("Removed %s from list", item.Name)
					}

					// Remove from map and list
					delete(m.Forwards, item.ID)
					idx := m.ForwardList.Index()
					if idx >= 0 {
						m.ForwardList.RemoveItem(idx)
					}
				}
			}
			return m, cmd
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
		switch m.ActiveTab {
		case 1:
			if m.Pane == 0 {
				m.Dialogs, cmd = m.Dialogs.Update(msg)
				return m, cmd
			} else {
				m.Messages, cmd = m.Messages.Update(msg)
				return m, cmd
			}
		case 2:
			var cmd tea.Cmd
			m.DownloadList, cmd = m.DownloadList.Update(msg)
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
		m.ForwardList.SetSize(m.width-2, listHeight)
		m.AccountsList.SetSize(m.width-10, listHeight-10) // Modals get slightly smaller boxes

		// Resize Inputs dynamically
		m.input.Width = m.width - 2
		m.ExportInput.Width = m.width - 2

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
		case "o":
			if m.DLForm.ActiveIndex == 0 {
				m.state = stateDirPicker
				// Initialize Directory Picker if needed
				// For now we reuse FilePicker but we might need a dir mode
				m.FilePicker.CurrentDirectory, _ = os.Getwd()
				return m, m.FilePicker.Init()
			}

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

func (m *Model) updateLogin(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.state = stateDashboard
			m.StatusMessage = "Login cancelled."
			return m, nil
		case "enter":
			switch m.state {
			case stateLoginPhone:
				phone := m.AuthPhone.Value()
				if phone != "" {
					m.state = stateLogin
					m.StatusMessage = "Looking up phone number..."
					return m, tea.Batch(m.loginSendCode(phone), m.spinner.Tick)
				}
			case stateLoginCode:
				code := m.AuthCode.Value()
				if code != "" {
					m.state = stateLogin
					m.StatusMessage = "Verifying code..."
					return m, tea.Batch(m.loginVerifyCode(code), m.spinner.Tick)
				}
			case stateLoginPassword:
				password := m.AuthPassword.Value()
				if password != "" {
					m.state = stateLogin
					m.StatusMessage = "Verifying password..."
					return m, tea.Batch(m.loginVerifyPassword(password), m.spinner.Tick)
				}
			}
		}
	}

	var cmd tea.Cmd
	switch m.state {
	case stateLoginPhone:
		m.AuthPhone, cmd = m.AuthPhone.Update(msg)
	case stateLoginCode:
		m.AuthCode, cmd = m.AuthCode.Update(msg)
	case stateLoginPassword:
		m.AuthPassword, cmd = m.AuthPassword.Update(msg)
	case stateLogin:
		m.spinner, cmd = m.spinner.Update(msg)
	}
	return m, cmd
}
