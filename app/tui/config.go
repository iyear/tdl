package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iyear/tdl/pkg/consts"
	"github.com/spf13/viper"
)

type configCategory struct {
	Name string
	Keys []string
}

var configLayout = []configCategory{
	{
		Name: "General & Network",
		Keys: []string{consts.FlagNamespace, consts.FlagProxy, "notify"},
	},
	{
		Name: "Downloading",
		Keys: []string{
			"download_dir", consts.FlagDlTemplate, "group", "skip_same",
			"takeout", "continue",
		},
	},
	{
		Name: "Advanced",
		Keys: []string{
			consts.FlagThreads, consts.FlagLimit, consts.FlagPartSize,
			consts.FlagDelay, consts.FlagReconnectTimeout,
		},
	},
	{
		Name: "Theme (Default, Catppuccin Macchiato, Dracula, Nord, Tokyo Night)",
		Keys: []string{
			"theme.name",
		},
	},
}

// Flatten keys for indexing
var configKeys []string

var configLabels = map[string]string{
	consts.FlagNamespace:        "Session Name",
	consts.FlagProxy:            "Proxy URL (e.g. socks5://)",
	"notify":                    "Enable Notifications (true/false)",
	"download_dir":              "Download Directory Path",
	consts.FlagDlTemplate:       "Filename Parsing Template",
	"group":                     "Group Media Sets (true/false)",
	"skip_same":                 "Skip Duplicates (true/false)",
	"takeout":                   "Use Takeout Session (true/false)",
	"continue":                  "Continue interrupted (true/false)",
	consts.FlagThreads:          "Concurrent Threads (number)",
	consts.FlagLimit:            "Download Limit in MB/s",
	consts.FlagPartSize:         "Part Chunk Size (Bytes)",
	consts.FlagDelay:            "Delay Between Files",
	consts.FlagReconnectTimeout: "Connection Retry Timeout",
	"theme.name":                "Active TUI Theme",
}

func init() {
	for _, cat := range configLayout {
		configKeys = append(configKeys, cat.Keys...)
	}
}

func (m *Model) InitConfigInputs() {
	m.ConfigInputs = make([]textinput.Model, len(configKeys))
	for i, key := range configKeys {
		t := textinput.New()
		t.Cursor.Style = lipgloss.NewStyle().Foreground(ColorPrimary)
		
		label := configLabels[key]
		if label == "" {
			label = key
		}
		t.Prompt = fmt.Sprintf("%-34s ", label+":")
		t.PromptStyle = lipgloss.NewStyle().Foreground(ColorSecondary)
		t.Width = 35

		val := viper.GetString(key)
		if key == "download_dir" && val == "" {
			val = "downloads"
		}
		t.SetValue(val)

		if i == 0 {
			t.Focus()
			t.PromptStyle = lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true)
		} else {
			t.Blur()
		}

		m.ConfigInputs[i] = t
	}
	m.ConfigFocusIndex = 0
}

func (m *Model) SaveConfig() error {
	for i, input := range m.ConfigInputs {
		key := configKeys[i]
		val := input.Value()
		
		switch val {
		case "true":
			viper.Set(key, true)
		case "false":
			viper.Set(key, false)
		default:
			if intVal, err := strconv.Atoi(val); err == nil {
				viper.Set(key, intVal)
			} else {
				viper.Set(key, val)
			}
		}
	}

	// Apply Theme immediately
	themeName := viper.GetString("theme.name")
	if themeName == "" {
		themeName = "Default"
	}
	ApplyTheme(themeName)

	if viper.ConfigFileUsed() != "" {
		return viper.WriteConfig()
	}
	return viper.WriteConfigAs("tdl.toml")
}

func (m *Model) updateConfig(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab", "up", "down":
			s := msg.String()

			if s == "up" || s == "shift+tab" {
				m.ConfigFocusIndex--
			} else {
				m.ConfigFocusIndex++
			}

			if m.ConfigFocusIndex > len(m.ConfigInputs) {
				m.ConfigFocusIndex = 0
			} else if m.ConfigFocusIndex < 0 {
				m.ConfigFocusIndex = len(m.ConfigInputs)
			}

			cmds := make([]tea.Cmd, len(m.ConfigInputs))
			for i := 0; i < len(m.ConfigInputs); i++ {
				if i == m.ConfigFocusIndex {
					cmds[i] = m.ConfigInputs[i].Focus()
					m.ConfigInputs[i].PromptStyle = lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true)
				} else {
					m.ConfigInputs[i].Blur()
					m.ConfigInputs[i].PromptStyle = lipgloss.NewStyle().Foreground(ColorSecondary)
				}
			}
			return m, tea.Batch(cmds...)

		case "enter":
			if m.ConfigFocusIndex == len(m.ConfigInputs) {
				if err := m.SaveConfig(); err != nil {
					m.StatusMessage = fmt.Sprintf("Error saving config: %v", err)
				} else {
					m.StatusMessage = "Configuration saved successfully."
				}
				m.state = stateDashboard
				return m, nil
			} else {
				// Move down
				return m.updateConfig(tea.KeyMsg{Type: tea.KeyDown})
			}

		case "esc":
			m.state = stateDashboard
			return m, nil
		}
	}

	cmd := m.updateInputs(msg)
	return m, cmd
}

func (m *Model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.ConfigInputs))
	for i := range m.ConfigInputs {
		m.ConfigInputs[i], cmds[i] = m.ConfigInputs[i].Update(msg)
	}
	return tea.Batch(cmds...)
}

func (m *Model) viewConfig() string {
	var s strings.Builder
	s.WriteString(TitleStyle.Render("Global Configuration Editor") + "\n\n")

	// Render by category
	catBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(ColorDim).
		Padding(0, 1).
		MarginBottom(1)

	catTitleStyle := lipgloss.NewStyle().Foreground(ColorSecondary).Bold(true)

	currInputIdx := 0

	// Create two columns conceptually, or just render top to bottom
	var blocks []string

	for _, cat := range configLayout {
		var catContent strings.Builder
		catContent.WriteString(catTitleStyle.Render(cat.Name) + "\n\n")

		for range cat.Keys {
			catContent.WriteString("  " + m.ConfigInputs[currInputIdx].View() + "\n")
			currInputIdx++
		}
		blocks = append(blocks, catBoxStyle.Render(catContent.String()))
	}

	// Layout in 2 columns
	var finalBlocks []string
	for i := 0; i < len(blocks); i += 2 {
		if i+1 < len(blocks) {
			row := lipgloss.JoinHorizontal(lipgloss.Top, blocks[i], "  ", blocks[i+1])
			finalBlocks = append(finalBlocks, row)
		} else {
			finalBlocks = append(finalBlocks, blocks[i])
		}
	}

	s.WriteString(lipgloss.JoinVertical(lipgloss.Left, finalBlocks...))
	s.WriteString("\n")

	// Save Button
	btn := "[ Save Configuration ]"
	if m.ConfigFocusIndex == len(m.ConfigInputs) {
		btn = lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true).Render(btn)
	} else {
		btn = InactivePaneStyle.Render(btn)
	}

	s.WriteString("  " + btn + "\n\n")
	s.WriteString(StatusBarStyle.Render("  [Tab/Arrows] Navigate Fields • [Enter] Save/Next • [Esc] Cancel"))

	return s.String()
}
