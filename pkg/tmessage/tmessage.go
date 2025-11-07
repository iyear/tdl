package tmessage

import (
	"github.com/gotd/td/tg"
)

// MessageMeta contains optional metadata about a message (from JSON export)
type MessageMeta struct {
	ID       int
	Filename string // Original filename from JSON export
}

type Dialog struct {
	Peer         tg.InputPeerClass
	Messages     []int
	MessageMetas map[int]*MessageMeta // Optional: metadata for messages from JSON (key is message ID)
}

type ParseSource func() ([]*Dialog, error)

func Parse(src ParseSource) ([]*Dialog, error) {
	return src()
}
