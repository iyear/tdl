package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	// Colors (Mutable)
	ColorPrimary   lipgloss.Color
	ColorSecondary lipgloss.Color
	ColorError     lipgloss.Color
	ColorSuccess   lipgloss.Color
	ColorDim       lipgloss.Color

	// Text Styles
	TitleStyle     lipgloss.Style
	StatusBarStyle lipgloss.Style

	// Item Styles
	SelectedItemStyle lipgloss.Style
	NormalItemStyle   lipgloss.Style

	// Pane Styles
	PaneStyle         lipgloss.Style
	ActivePaneStyle   lipgloss.Style
	InactivePaneStyle lipgloss.Style

	// Tab Styles
	TabStyle         lipgloss.Style
	ActiveTabStyle   lipgloss.Style
	InactiveTabStyle lipgloss.Style
)

type Theme struct {
	Primary   string
	Secondary string
	Error     string
	Success   string
	Dim       string
}

var Themes = map[string]Theme{
	"Default": {
		Primary:   "62",
		Secondary: "230",
		Error:     "196",
		Success:   "42",
		Dim:       "240",
	},
	"Catppuccin Macchiato": {
		Primary:   "#8aadf4",
		Secondary: "#c6a0f6",
		Error:     "#ed8796",
		Success:   "#a6da95",
		Dim:       "#5b6078",
	},
	"Dracula": {
		Primary:   "#bd93f9",
		Secondary: "#ff79c6",
		Error:     "#ff5555",
		Success:   "#50fa7b",
		Dim:       "#6272a4",
	},
	"Nord": {
		Primary:   "#88c0d0",
		Secondary: "#b48ead",
		Error:     "#bf616a",
		Success:   "#a3be8c",
		Dim:       "#4c566a",
	},
	"Tokyo Night": {
		Primary:   "#7aa2f7",
		Secondary: "#bb9af7",
		Error:     "#f7768e",
		Success:   "#9ece6a",
		Dim:       "#565f89",
	},
}

func init() {
	ApplyTheme("Default")
}

func ApplyTheme(name string) {
	theme, ok := Themes[name]
	if !ok {
		theme = Themes["Default"]
	}

	ColorPrimary = lipgloss.Color(theme.Primary)
	ColorSecondary = lipgloss.Color(theme.Secondary)
	ColorError = lipgloss.Color(theme.Error)
	ColorSuccess = lipgloss.Color(theme.Success)
	ColorDim = lipgloss.Color(theme.Dim)

	TitleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("230")).
		Background(ColorPrimary).
		Bold(true).
		Padding(0, 2)

	StatusBarStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("230")).
		Background(ColorDim).
		Padding(0, 1)

	SelectedItemStyle = lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Bold(true).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(ColorPrimary).
		PaddingLeft(1)

	NormalItemStyle = lipgloss.NewStyle().
		PaddingLeft(2)

	PaneStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder())

	ActivePaneStyle = PaneStyle.Copy().
		BorderForeground(ColorPrimary)

	InactivePaneStyle = PaneStyle.Copy().
		BorderForeground(ColorDim)

	TabStyle = lipgloss.NewStyle().
		Padding(0, 2)

	ActiveTabStyle = TabStyle.Copy().
		Foreground(lipgloss.Color("230")).
		Background(ColorPrimary).
		Bold(true)

	InactiveTabStyle = TabStyle.Copy().
		Foreground(ColorDim).
		Background(lipgloss.Color("236"))
}

// Icons (Nerd Font friendly/Unicode)
const (
	IconFolder   = "📁"
	IconFile     = "📄"
	IconPhoto    = "🖼️"
	IconVideo    = "🎥"
	IconMusic    = "🎵"
	IconUnknown  = "❓"
	IconCheck    = "✅"
	IconError    = "❌"
	IconDownload = "⬇️"
	IconWaiting  = "⏳"
)
