package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m *Model) View() string {
	if m.quitting {
		return "Bye!\n"
	}

	var s string

	// Header
	header := TitleStyle.Render("TDL TUI")
	status := lipgloss.NewStyle().Foreground(ColorError).Render("Disconnected")
	if m.Connected {
		status = lipgloss.NewStyle().Foreground(ColorSuccess).Render("Connected")
	}

	s += lipgloss.JoinHorizontal(lipgloss.Center, header, "  ", status)
	s += "\n\n"

	// Tabs
	s += m.viewTabs()
	s += "\n\n"

	// Main Content
	// Handle different tabs
	switch m.state {
	case stateConfig:
		s += m.viewConfig()
	case stateBatch:
		s += m.viewBatch()
	case stateBatchConfirm:
		s += m.viewBatchConfirm()
	case stateDownloads:
		s += m.viewDownloads()
	case stateExportPrompt:
		s += m.viewExportPrompt()
	case stateDirPicker:
		s += m.viewDirPicker()
	case stateDownloadOptions:
		s += m.viewDownloadOptions()
	case stateAccounts:
		s += m.viewAccounts()
	case stateLogin, stateLoginPhone, stateLoginCode, stateLoginPassword:
		s += m.viewLogin()
	default:
		// ActiveTab handling when on dashboard/browser
		switch m.ActiveTab {
		case 1:
			s += m.viewBrowser()
		case 2:
			s += m.viewDownloads()
		case 3:
			s += m.viewForwarding()
		default:
			s += m.viewDashboard()
		}
	}

	if m.ShowHelp {
		return m.viewHelpModal()
	}

	return s
}

func (m *Model) viewDownloadOptions() string {
	var s strings.Builder
	s.WriteString(TitleStyle.Render("Download Configuration"))
	s.WriteString("\n\n")

	s.WriteString(fmt.Sprintf("Target: %s\n\n", m.DLForm.UrlOrPath))

	// Helpers for form rendering
	focused := func(idx int) lipgloss.Style {
		if m.DLForm.ActiveIndex == idx {
			return lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true)
		}
		return lipgloss.NewStyle().Foreground(ColorDim)
	}

	checkbox := func(label string, checked bool, idx int) string {
		icon := "[ ]"
		if checked {
			icon = "[x]"
		}
		style := focused(idx)
		return style.Render(fmt.Sprintf("%s %s", icon, label))
	}

	// 0: Dir
	s.WriteString(focused(0).Render("Directory: (Press 'o' to use picker)") + "\n")
	s.WriteString(m.DLForm.Dir.View() + "\n\n")

	// 1: Template
	s.WriteString(focused(1).Render("Filename Template:") + "\n")
	s.WriteString(m.DLForm.Template.View() + "\n\n")

	// Options Grid (2-5)
	// Group | SkipSame | Takeout | Desc
	row1 := lipgloss.JoinHorizontal(lipgloss.Top,
		checkbox("Group Media", m.DLForm.Group, 2), "  ",
		checkbox("Skip Duplicates", m.DLForm.SkipSame, 3))

	row2 := lipgloss.JoinHorizontal(lipgloss.Top,
		checkbox("Takeout Session", m.DLForm.Takeout, 4), " ", // slightly less padding to fit
		checkbox("Reverse Order", m.DLForm.Desc, 5))

	s.WriteString(row1 + "\n" + row2 + "\n\n")

	s.WriteString(lipgloss.NewStyle().Foreground(ColorSecondary).Bold(true).Render("Advanced Options") + "\n")

	// Advanced Inputs (6-10)
	// Threads | Limit | Pool
	// Delay   | Reconnect

	label := func(text string, idx int) string {
		return focused(idx).Render(text)
	}

	advRow1 := lipgloss.JoinHorizontal(lipgloss.Top,
		label("Threads: ", 6)+m.DLForm.Threads.View(), "  ",
		label("Limit: ", 7)+m.DLForm.Limit.View(), "  ",
		label("Pool: ", 8)+m.DLForm.Pool.View())

	advRow2 := lipgloss.JoinHorizontal(lipgloss.Top,
		label("Delay: ", 9)+m.DLForm.Delay.View(), "  ",
		label("Reconnect: ", 10)+m.DLForm.Reconnect.View())

	s.WriteString(advRow1 + "\n" + advRow2 + "\n\n")

	// Advanced Bools (11-12)
	advRow3 := lipgloss.JoinHorizontal(lipgloss.Top,
		checkbox("Continue", m.DLForm.Continue, 11), "  ",
		checkbox("Debug Log", m.DLForm.Debug, 12))

	s.WriteString(advRow3 + "\n\n")

	// Buttons (13-14)
	btnStart := "[ Start Download ]"
	if m.DLForm.ActiveIndex == 13 {
		btnStart = ActiveTabStyle.Render("[ Start Download ]")
	} else {
		btnStart = InactiveTabStyle.Render("[ Start Download ]")
	}

	btnCancel := "[ Cancel ]"
	if m.DLForm.ActiveIndex == 14 {
		btnCancel = ActiveTabStyle.Render("[ Cancel ]")
	} else {
		btnCancel = InactiveTabStyle.Render("[ Cancel ]")
	}

	s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, btnStart, "  ", btnCancel))

	s.WriteString("\n\n")
	s.WriteString(StatusBarStyle.Render("[Tab/Arrows] Navigate • [Enter] Toggle/Select • [Esc] Cancel"))

	return s.String()
}

func (m *Model) viewHelpModal() string {
	// Simple centered box
	keys := []string{
		"Navigation ------------------------",
		"  Tab       Switch between Panes",
		"  Arrows    Navigate Lists & Options",
		"  Enter     Select Item / Open / Toggle",
		"  Esc       Go Back / Blur Input",
		"",
		"General ---------------------------",
		"  d         Go to Dashboard",
		"  b         Go to Browser",
		"  l         Go to Downloads",
		"  ?         Toggle Help",
		"  q         Quit Application",
		"",
		"Browser Actions -------------------",
		"  Space     Select/Deselect Message",
		"  /         Filter Chats or Search",
		"  f         Forward Selected Messages",
		"  e         Export Chat Info (JSON)",
		"  L         Load More Chats",
		"",
		"Download Actions ------------------",
		"  i         New Download (URL -> Configure)",
		"  j         Batch Download (JSON -> Configure)",
		"  a         Switch Account",
	}

	content := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPrimary).
		Padding(1, 2).
		Render(strings.Join(keys, "\n"))

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		content,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("237")),
	)
}

func (m *Model) viewBrowser() string {
	var s string

	// Layout Calculations
	availableWidth := m.width - 5
	if availableWidth < 0 {
		availableWidth = 0
	}

	leftWidth := availableWidth / 3
	rightWidth := availableWidth - leftWidth

	listHeight := m.height - 10
	if listHeight < 0 {
		listHeight = 0
	}

	// Left Pane (Dialogs)
	leftStyle := InactivePaneStyle.Copy().
		Width(leftWidth).
		Height(listHeight)

	if m.PickingDest {
		leftStyle = ActivePaneStyle.Copy().
			Width(leftWidth).
			Height(listHeight).
			BorderForeground(lipgloss.Color("205")) // Pink for special mode
	}

	// Left Content
	var leftContent string
	if m.LoadingDialogs {
		leftContent = fmt.Sprintf("\n\n   %s Loading chats...", m.spinner.View())
	} else {
		leftContent = m.Dialogs.View()
	}
	left := leftStyle.Render(leftContent)

	// Right Pane (Messages)
	// ... (Right pane style logic same as before)
	rightStyle := InactivePaneStyle.Copy().
		Width(rightWidth).
		Height(listHeight).
		MarginLeft(1)

	if m.Pane == 1 {
		rightStyle = ActivePaneStyle.Copy().
			Width(rightWidth).
			Height(listHeight).
			MarginLeft(1)
	}

	// Right Content
	var rightContent string
	if m.LoadingHistory {
		rightContent = fmt.Sprintf("\n\n   %s Loading messages...", m.spinner.View())
	} else if len(m.Messages.Items()) > 0 {
		rightContent = m.Messages.View()
	} else {
		rightContent = "Select a chat to view messages"
	}
	right := rightStyle.Render(rightContent)

	s = lipgloss.JoinHorizontal(lipgloss.Top, left, right)

	if m.PickingDest {
		s += StatusBarStyle.Render("\n  [Enter] Confirm Destination (Forward) • [Esc] Cancel")
	} else {
		s += StatusBarStyle.Render("\n  [Tab] Pane • [Enter] Select • [Space] Mark • [e] Export • [?] Help")
	}

	if m.StatusMessage != "" {
		color := ColorSuccess
		if m.LoadingExport {
			color = ColorPrimary
		}
		s += lipgloss.NewStyle().Foreground(color).Render("\n  " + m.StatusMessage)
	} else if m.LoadingExport {
		s += lipgloss.NewStyle().Foreground(ColorPrimary).Render("\n  ⏳ Exporting chat info... This may take a while.")
	} else if !m.PickingDest {
		// Show selection count
		count := 0
		for _, item := range m.Messages.Items() {
			if mItem, ok := item.(MessageItem); ok && mItem.Selected {
				count++
			}
		}
		if count > 0 {
			s += lipgloss.NewStyle().Foreground(ColorPrimary).Render(fmt.Sprintf("\n  %d messages selected", count))
		}
	}

	if m.Searching || (m.ActiveTab == 1 && m.input.Focused()) {
		s += "\n\n  " + m.input.View()
	}

	return s
}

func (m *Model) viewDashboard() string {
	var s strings.Builder

	// Logo
	logo := `
████████╗██████╗ ██╗
╚══██╔══╝██╔══██╗██║
   ██║   ██║  ██║██║
   ██║   ██║  ██║██║
   ██║   ██████╔╝███████╗
   ╚═╝   ╚═════╝ ╚══════╝`

	s.WriteString(lipgloss.NewStyle().Foreground(ColorPrimary).Render(logo))
	s.WriteString("\n\n")

	if m.Connected {
		s.WriteString(lipgloss.NewStyle().Foreground(ColorSuccess).Render("  You are connected to Telegram."))
		if m.User != nil {
			user := fmt.Sprintf("\n  User: %s %s (@%s)\n  ID: %d",
				m.User.FirstName, m.User.LastName, m.User.Username, m.User.ID)
			s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render(user))
		}
	} else {
		s.WriteString(lipgloss.NewStyle().Foreground(ColorError).Render("  Not connected."))
		s.WriteString("\n  Please login via 'tdl login' first or check your configuration.\n")
	}

	// Accounts Section
	if len(m.Accounts) > 1 {
		s.WriteString("\n\n" + TitleStyle.Render("Accounts [a]:"))
		for _, acc := range m.Accounts {
			style := lipgloss.NewStyle().Foreground(ColorSecondary)
			prefix := "  "
			if acc == m.Namespace {
				style = lipgloss.NewStyle().Foreground(ColorSuccess).Bold(true)
				prefix = "> "
			}
			s.WriteString("\n" + style.Render(fmt.Sprintf("%s%s", prefix, acc)))
		}
	}

	// System Metrics Section
	s.WriteString("\n\n" + TitleStyle.Render("System Health:"))
	memStr := fmt.Sprintf("RAM: %.1f%%", m.sysMem)
	cpuStr := fmt.Sprintf("CPU: %.1f%%", m.sysCpu)
	s.WriteString(lipgloss.NewStyle().Foreground(ColorSecondary).Render("\n  " + cpuStr + "  •  " + memStr))

	s.WriteString("\n\n  [d] Dashboard  [b] Browser  [l] Downloads  [c] Config  [j] Batch  [i] New Download  [r] Login  [q] Quit")

	if m.input.Focused() {
		s.WriteString("\n\n")
		s.WriteString(m.input.View())
	}

	// Footer
	s.WriteString("\n\n")
	if m.StatusMessage != "" {
		s.WriteString(lipgloss.NewStyle().Foreground(ColorPrimary).Render("  " + m.StatusMessage + "\n"))
	}
	s.WriteString(StatusBarStyle.Render(fmt.Sprintf("tdl %s • %s • [j] Batch • [?] Help", m.BuildInfo, m.Namespace)))

	return s.String()
}

func (m *Model) viewForwarding() string {
	var s strings.Builder
	s.WriteString("Active Forwarding Clones:\n\n")

	if len(m.Forwards) == 0 {
		s.WriteString("  No active forwards.\n\n")
		s.WriteString(lipgloss.NewStyle().Foreground(ColorDim).Render("  Select messages in a chat and press 'f', or press 'j' to load a JSON export."))
	} else {
		s.WriteString(m.ForwardList.View())
	}
	return s.String()
}

func (m *Model) viewAccounts() string {
	var s string

	s += TitleStyle.Render("Session Manager") + "\n\n"
	s += "  Active Session: " + lipgloss.NewStyle().Foreground(ColorSuccess).Render(m.Namespace) + "\n\n"

	s += m.AccountsList.View() + "\n\n"

	s += StatusBarStyle.Render("  [Enter] Switch Session • [n] New Session • [Esc] Close")

	return s
}

func (m *Model) viewDownloads() string {
	var s strings.Builder
	s.WriteString("Active Downloads:\n\n")

	if len(m.Downloads) == 0 {
		s.WriteString("  No active downloads.\n\n")
		s.WriteString(lipgloss.NewStyle().Foreground(ColorDim).Render("  Press 'i' to start a new download from a URL."))
		s.WriteString(lipgloss.NewStyle().Foreground(ColorDim).Render("\n  Press 'j' to start a Batch Download (JSON)."))

		if m.input.Focused() {
			s.WriteString("\n\n  " + m.input.View())
		}

		return s.String()
	}

	// Render the interactive list
	s.WriteString(m.DownloadList.View())

	if m.input.Focused() {
		s.WriteString("\n\n  " + m.input.View())
	} else {
		s.WriteString(StatusBarStyle.Render("\n  [Arrows] Navigate • [o] Open • [i] New • [j] Batch • [x] Cancel"))
	}

	return s.String()
}

func (m *Model) viewBatch() string {
	var s strings.Builder
	s.WriteString(TitleStyle.Render("Batch Download (JSON)"))
	s.WriteString("\n\n  Select a JSON file containing message/media data:\n\n")
	s.WriteString(m.FilePicker.View() + "\n")
	s.WriteString(StatusBarStyle.Render("\n  [Esc] Back • [Enter] Select Directory/File"))
	return s.String()
}

func (m *Model) viewBatchConfirm() string {
	var s strings.Builder
	s.WriteString(TitleStyle.Render("Confirm Batch Action"))
	s.WriteString("\n\n")
	s.WriteString(fmt.Sprintf("Selected File: %s", m.BatchPath))
	s.WriteString("\n\n")
	s.WriteString("What would you like to do?\n\n")
	s.WriteString(lipgloss.NewStyle().Foreground(ColorSuccess).Render("[d] Download (Default)"))
	s.WriteString("\n")
	s.WriteString(lipgloss.NewStyle().Foreground(ColorSecondary).Render("[f] Forward to Chat"))
	s.WriteString("\n\n")
	s.WriteString(StatusBarStyle.Render("[Esc] Cancel"))
	return s.String()
}

func (m *Model) viewTabs() string {
	var tabs []string
	// Definition of tabs: 0=Dashboard, 1=Browser, 2=Downloads, 3=Forwarding
	labels := []string{"Dashboard", "Browser", "Downloads", "Forwarding"}

	for i, label := range labels {
		style := InactiveTabStyle
		if m.ActiveTab == i {
			style = ActiveTabStyle
		}
		tabs = append(tabs, style.Render(label))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
}

func (m *Model) viewExportPrompt() string {
	var s strings.Builder
	s.WriteString(TitleStyle.Render("Export Chat Data"))
	s.WriteString("\n\n")
	s.WriteString("Enter filename for export (JSON):\n")
	s.WriteString(m.ExportInput.View())
	s.WriteString("\n\n")
	s.WriteString(StatusBarStyle.Render("[Enter] Export • [Esc] Cancel"))
	return s.String()
}

func (m *Model) viewLogin() string {
	var s strings.Builder

	s.WriteString(TitleStyle.Render("Telegram Login"))
	s.WriteString("\n\n")

	if m.StatusMessage != "" {
		s.WriteString(lipgloss.NewStyle().Foreground(ColorPrimary).Render(m.StatusMessage))
		s.WriteString("\n\n")
	} else if m.state == stateLogin {
		s.WriteString(lipgloss.NewStyle().Foreground(ColorPrimary).Render(fmt.Sprintf("\n\n   %s Initializing login...", m.spinner.View())))
		s.WriteString("\n\n")
	}

	switch m.state {
	case stateLoginPhone:
		s.WriteString("Enter your phone number (with country code):")
		s.WriteString("\n\n")
		s.WriteString(m.AuthPhone.View())
	case stateLoginCode:
		s.WriteString("We've sent a code to the Telegram app on your other device.")
		s.WriteString("\nPlease enter the code below:")
		s.WriteString("\n\n")
		s.WriteString(m.AuthCode.View())
	case stateLoginPassword:
		s.WriteString("Your account is protected with a 2-Step Verification password.")
		s.WriteString("\nPlease enter your password:")
		s.WriteString("\n\n")
		s.WriteString(m.AuthPassword.View())
	}

	s.WriteString("\n\n")
	s.WriteString(StatusBarStyle.Render("[Enter] Submit • [Esc] Cancel"))

	return s.String()
}

func (m *Model) viewDirPicker() string {
	var s strings.Builder
	s.WriteString(TitleStyle.Render("Select Download Directory"))
	s.WriteString("\n\n  Choose a destination folder:\n\n")
	s.WriteString(m.FilePicker.View() + "\n")
	s.WriteString(StatusBarStyle.Render("\n  [Esc] Cancel • [Enter] Navigate • [s] Select Current Directory"))
	return s.String()
}
