package tmessage

import (
	"github.com/gotd/td/tg"
)

// MessageMeta contains optional metadata about a message (from JSON export)
type MessageMeta struct {
	ID          int
	Filename    string // Original filename from JSON export (file or photo field)
	Date        int64  // Message date as unix timestamp
	TextContent string // Message text/caption content
}

type Dialog struct {
	Peer         tg.InputPeerClass
	Messages     []int
	MessageMetas map[int]*MessageMeta // Optional: metadata for messages from JSON (key is message ID)
	HasRawData   bool                 // True if JSON export includes raw Telegram message data
}

type ParseSource func() ([]*Dialog, error)

func Parse(src ParseSource) ([]*Dialog, error) {
	return src()
}
