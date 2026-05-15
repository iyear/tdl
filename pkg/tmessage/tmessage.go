package tmessage

import (
	"github.com/gotd/td/tg"
)

// TextMsg holds the essential data of a text-only message (no downloadable media)
// for later export as HTML.
type TextMsg struct {
	ID           int
	Date         int
	Text         string
	Entities     []tg.MessageEntityClass
	ReplyToMsgID int
	FromName     string
}

type Dialog struct {
	Peer         tg.InputPeerClass
	Messages     []int
	TextMessages []TextMsg // text-only messages (for HTML export)
}

type ParseSource func() ([]*Dialog, error)

func Parse(src ParseSource) ([]*Dialog, error) {
	return src()
}

// GetDialogPeerID extracts the numeric ID from an InputPeerClass.
func GetDialogPeerID(peer tg.InputPeerClass) int64 {
	switch p := peer.(type) {
	case *tg.InputPeerUser:
		return p.UserID
	case *tg.InputPeerChat:
		return p.ChatID
	case *tg.InputPeerChannel:
		return p.ChannelID
	}
	return 0
}
