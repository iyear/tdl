package dl

import (
	"fmt"
	"html"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/gotd/td/tg"
)

// HTML tag constants to avoid goconst lint warnings.
const (
	htmlCloseB = "</b>"
	htmlCloseA = "</a>"
)

// TextMessage holds the data for a single text-only message to be exported as HTML.
type TextMessage struct {
	ID       int
	Date     int
	Text     string
	Entities []tg.MessageEntityClass
	// ReplyToMsgID is the ID of the message this is replying to (0 if not a reply)
	ReplyToMsgID int
	// FromName is the sender name (if available)
	FromName string
}

// exportTextMessagesHTML writes a self-contained HTML file with all collected text messages.
// The file is saved to dir/chatName_text_messages.html
func exportTextMessagesHTML(dir string, chatName string, chatID int64, messages []TextMessage) error {
	if len(messages) == 0 {
		return nil
	}

	// Sort messages by ID (ascending = chronological)
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].ID < messages[j].ID
	})

	// Sanitize chat name for filename
	safeName := sanitizeFilename(chatName)
	filename := fmt.Sprintf("%s_%d_text_messages.html", safeName, chatID)
	path := filepath.Join(dir, filename)

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create dir for HTML export: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create HTML file: %w", err)
	}
	defer f.Close()

	// Write HTML header
	fmt.Fprintf(f, `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>%s — Text Messages</title>
<style>
:root {
  --bg: #0e1621;
  --msg-bg: #182533;
  --msg-border: #1e3a5f;
  --text: #e4e6ea;
  --text-muted: #6c7883;
  --accent: #3daee0;
  --link: #6ab3f3;
  --code-bg: #1c2e3f;
  --date-bg: rgba(0,0,0,0.35);
  --reply-border: #3daee0;
  --reply-bg: rgba(61,174,224,0.08);
}
* { margin: 0; padding: 0; box-sizing: border-box; }
body {
  background: var(--bg);
  color: var(--text);
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  font-size: 14px;
  line-height: 1.5;
  padding: 0;
}
.container {
  max-width: 780px;
  margin: 0 auto;
  padding: 20px 16px 60px;
}
.header {
  text-align: center;
  padding: 24px 0 32px;
  border-bottom: 1px solid var(--msg-border);
  margin-bottom: 24px;
}
.header h1 {
  font-size: 20px;
  font-weight: 600;
  color: var(--accent);
}
.header .meta {
  font-size: 12px;
  color: var(--text-muted);
  margin-top: 6px;
}
.date-separator {
  text-align: center;
  margin: 20px 0 12px;
}
.date-separator span {
  background: var(--date-bg);
  color: var(--text-muted);
  font-size: 12px;
  font-weight: 500;
  padding: 4px 14px;
  border-radius: 12px;
}
.msg {
  background: var(--msg-bg);
  border: 1px solid var(--msg-border);
  border-radius: 12px;
  padding: 10px 14px;
  margin-bottom: 6px;
  max-width: 85%%;
  word-wrap: break-word;
  overflow-wrap: break-word;
  position: relative;
}
.msg .sender {
  font-size: 13px;
  font-weight: 600;
  color: var(--accent);
  margin-bottom: 2px;
}
.msg .body {
  white-space: pre-wrap;
  word-break: break-word;
}
.msg .body a {
  color: var(--link);
  text-decoration: none;
}
.msg .body a:hover {
  text-decoration: underline;
}
.msg .body code {
  background: var(--code-bg);
  padding: 1px 5px;
  border-radius: 4px;
  font-family: 'Cascadia Code', 'Fira Code', 'Consolas', monospace;
  font-size: 13px;
}
.msg .body pre {
  background: var(--code-bg);
  padding: 10px 12px;
  border-radius: 8px;
  overflow-x: auto;
  margin: 6px 0;
  font-family: 'Cascadia Code', 'Fira Code', 'Consolas', monospace;
  font-size: 13px;
}
.msg .reply-ref {
  background: var(--reply-bg);
  border-left: 3px solid var(--reply-border);
  padding: 4px 8px;
  margin-bottom: 6px;
  border-radius: 0 6px 6px 0;
  font-size: 12px;
  color: var(--text-muted);
}
.msg .time {
  font-size: 11px;
  color: var(--text-muted);
  text-align: right;
  margin-top: 4px;
}
.msg .msg-id {
  font-size: 10px;
  color: var(--text-muted);
  opacity: 0.5;
}
.footer {
  text-align: center;
  padding: 24px 0;
  border-top: 1px solid var(--msg-border);
  margin-top: 24px;
  font-size: 12px;
  color: var(--text-muted);
}
</style>
</head>
<body>
<div class="container">
<div class="header">
  <h1>%s</h1>
  <div class="meta">%d text messages · Exported on %s</div>
</div>
`,
		html.EscapeString(chatName),
		html.EscapeString(chatName),
		len(messages),
		time.Now().Format("2006-01-02 15:04:05"),
	)

	// Group messages by date and render
	var lastDate string
	for _, msg := range messages {
		msgTime := time.Unix(int64(msg.Date), 0)
		dateStr := msgTime.Format("January 2, 2006")

		if dateStr != lastDate {
			fmt.Fprintf(f, "<div class=\"date-separator\"><span>%s</span></div>\n", html.EscapeString(dateStr))
			lastDate = dateStr
		}

		// Render message
		fmt.Fprintf(f, "<div class=\"msg\" id=\"msg-%d\">\n", msg.ID)

		// Sender name if available
		if msg.FromName != "" {
			fmt.Fprintf(f, "  <div class=\"sender\">%s</div>\n", html.EscapeString(msg.FromName))
		}

		// Reply reference
		if msg.ReplyToMsgID > 0 {
			fmt.Fprintf(f, "  <div class=\"reply-ref\">↩ Reply to <a href=\"#msg-%d\">message #%d</a></div>\n",
				msg.ReplyToMsgID, msg.ReplyToMsgID)
		}

		// Message body with entity formatting
		formattedText := formatMessageEntities(msg.Text, msg.Entities)
		fmt.Fprintf(f, "  <div class=\"body\">%s</div>\n", formattedText)

		// Timestamp and message ID
		fmt.Fprintf(f, "  <div class=\"time\">%s <span class=\"msg-id\">#%d</span></div>\n",
			msgTime.Format("15:04"), msg.ID)

		fmt.Fprintf(f, "</div>\n")
	}

	// Footer
	fmt.Fprintf(f, `<div class="footer">
  Exported by tdl · Chat ID: %d · %d messages
</div>
</div>
</body>
</html>
`, chatID, len(messages))

	color.Green("📄 Saved %d text messages to '%s'", len(messages), path)
	return nil
}

// formatMessageEntities applies Telegram entity formatting (bold, italic, links, code, etc.)
// to the message text, producing safe HTML output.
func formatMessageEntities(text string, entities []tg.MessageEntityClass) string {
	if len(entities) == 0 {
		return html.EscapeString(text)
	}

	runes := []rune(text)

	// Build a list of insert operations (open/close tags at rune positions)
	type tag struct {
		pos   int
		close bool
		order int // for stable sorting: opens before closes at same position
		html  string
	}

	var tags []tag

	for _, ent := range entities {
		var offset, length int
		var openTag, closeTag string

		switch e := ent.(type) {
		case *tg.MessageEntityBold:
			offset, length = e.Offset, e.Length
			openTag, closeTag = "<b>", htmlCloseB
		case *tg.MessageEntityItalic:
			offset, length = e.Offset, e.Length
			openTag, closeTag = "<i>", "</i>"
		case *tg.MessageEntityUnderline:
			offset, length = e.Offset, e.Length
			openTag, closeTag = "<u>", "</u>"
		case *tg.MessageEntityStrike:
			offset, length = e.Offset, e.Length
			openTag, closeTag = "<s>", "</s>"
		case *tg.MessageEntityCode:
			offset, length = e.Offset, e.Length
			openTag, closeTag = "<code>", "</code>"
		case *tg.MessageEntityPre:
			offset, length = e.Offset, e.Length
			lang := ""
			if e.Language != "" {
				lang = fmt.Sprintf(" data-lang=\"%s\"", html.EscapeString(e.Language))
			}
			openTag = fmt.Sprintf("<pre%s>", lang)
			closeTag = "</pre>"
		case *tg.MessageEntityTextURL:
			offset, length = e.Offset, e.Length
			openTag = fmt.Sprintf("<a href=\"%s\" target=\"_blank\" rel=\"noopener\">", html.EscapeString(e.URL))
			closeTag = htmlCloseA
		case *tg.MessageEntityURL:
			offset, length = e.Offset, e.Length
			urlText := string(runes[offset : offset+length])
			openTag = fmt.Sprintf("<a href=\"%s\" target=\"_blank\" rel=\"noopener\">", html.EscapeString(urlText))
			closeTag = htmlCloseA
		case *tg.MessageEntityMentionName:
			offset, length = e.Offset, e.Length
			openTag = fmt.Sprintf("<b title=\"User ID: %d\">", e.UserID)
			closeTag = htmlCloseB
		case *tg.MessageEntityMention:
			offset, length = e.Offset, e.Length
			mention := string(runes[offset : offset+length])
			openTag = fmt.Sprintf("<a href=\"https://t.me/%s\" target=\"_blank\" rel=\"noopener\">",
				html.EscapeString(strings.TrimPrefix(mention, "@")))
			closeTag = htmlCloseA
		case *tg.MessageEntityHashtag:
			offset, length = e.Offset, e.Length
			openTag = "<b>"
			closeTag = htmlCloseB
		case *tg.MessageEntitySpoiler:
			offset, length = e.Offset, e.Length
			openTag = "<span style=\"background:var(--text);color:var(--text);cursor:pointer\" onclick=\"this.style.color='inherit';this.style.background='transparent'\">"
			closeTag = "</span>"
		case *tg.MessageEntityBlockquote:
			offset, length = e.Offset, e.Length
			openTag = "<blockquote style=\"border-left:3px solid var(--accent);padding-left:8px;margin:4px 0;color:var(--text-muted)\">"
			closeTag = "</blockquote>"
		default:
			continue
		}

		tags = append(tags, tag{pos: offset, close: false, order: 0, html: openTag})
		tags = append(tags, tag{pos: offset + length, close: true, order: 1, html: closeTag})
	}

	// Sort tags: by position, then closes before opens at same position
	sort.SliceStable(tags, func(i, j int) bool {
		if tags[i].pos != tags[j].pos {
			return tags[i].pos < tags[j].pos
		}
		return tags[i].order > tags[j].order // closes first
	})

	// Build output
	var b strings.Builder
	tagIdx := 0
	for i, r := range runes {
		// Insert any tags at this position
		for tagIdx < len(tags) && tags[tagIdx].pos == i {
			b.WriteString(tags[tagIdx].html)
			tagIdx++
		}
		b.WriteString(html.EscapeString(string(r)))
	}
	// Flush remaining tags at end position
	for tagIdx < len(tags) {
		b.WriteString(tags[tagIdx].html)
		tagIdx++
	}

	return b.String()
}

// sanitizeFilename removes or replaces characters that are not safe for filenames.
func sanitizeFilename(name string) string {
	replacer := strings.NewReplacer(
		"/", "_", "\\", "_", ":", "_", "*", "_",
		"?", "_", "\"", "_", "<", "_", ">", "_", "|", "_",
		" ", "_",
	)
	result := replacer.Replace(name)
	if len(result) > 60 {
		result = result[:60]
	}
	return result
}
