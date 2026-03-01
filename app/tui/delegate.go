package tui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// ItemDelegate handles rendering of list items
type ItemDelegate struct{}

func (d ItemDelegate) Height() int                             { return 1 }
func (d ItemDelegate) Spacing() int                            { return 0 }
func (d ItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d ItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	var str string

	fn := NormalItemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return SelectedItemStyle.Render("> " + s[0])
		}
	}

	// Custom rendering based on type
	switch item := listItem.(type) {
	case DialogItem:
		icon := IconFolder
		title := item.Title
		// truncate title
		if len(title) > 20 {
			title = title[:17] + "..."
		}

		// desc := fmt.Sprintf("ID: %d", item.PeerID)
		if item.Unread > 0 {
			icon = "🔴" // Alert
		}

		str = fmt.Sprintf("%s %s", icon, title)
		if index == m.Index() {
			str = fmt.Sprintf("%s %s %s", ">", icon, title)
		}

		fmt.Fprint(w, fn(str))

	case MessageItem:
		icon := "💬" // IconMessage
		if item.HasMedia {
			switch item.Media {
			case "Photo":
				icon = IconPhoto
			case "Document":
				icon = IconFile
			default:
				icon = IconUnknown
			}
		}

		text := item.Text
		if text == "" {
			text = "[" + item.Media + "]"
		}

		// truncate
		if len(text) > 40 {
			text = text[:37] + "..."
		}

		str = fmt.Sprintf("%s %s", icon, text)
		if index == m.Index() {
			str = fmt.Sprintf("%s %s %s", ">", icon, text)
		}

		fmt.Fprint(w, fn(str))
	case *DownloadItem:
		str = fmt.Sprintf("%s  %s", item.Title(), item.Description())
		if index == m.Index() {
			str = "> " + str
		}
		fmt.Fprint(w, fn(str))
	}
}
