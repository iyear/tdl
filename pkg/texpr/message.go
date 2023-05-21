package texpr

import (
	"github.com/gotd/td/tg"
	"github.com/iyear/tdl/pkg/tmedia"
	"github.com/iyear/tdl/pkg/utils"
)

type Message struct {
	// Whether we were mentioned in this message
	Mentioned bool
	// Whether this is a silent message (no notification triggered)
	Silent bool
	// Whether this is a scheduled message
	FromScheduled bool
	// Whether this message is pinned
	Pinned bool
	// ID of the message
	ID int
	// ID of the sender of the message
	FromID int64
	// Date of the message
	Date int
	// The message
	Message string
	// Media attachment
	Media MessageMedia
	// View count
	Views int
	// Forward count
	Forwards int
}

type MessageMedia struct {
	// File name
	Name string
	// File size. Unit: Byte
	Size int64
	// DC ID
	DC int
}

func CovertMessage(msg *tg.Message) (m *Message) {
	m = &Message{}

	m.Mentioned = msg.Mentioned
	m.Silent = msg.Silent
	m.FromScheduled = msg.FromScheduled
	m.Pinned = msg.Pinned
	m.ID = msg.ID
	m.FromID = utils.Telegram.GetPeerID(msg.FromID)
	m.Date = msg.Date
	m.Message = msg.Message

	if media, ok := tmedia.GetMedia(msg); ok {
		m.Media = MessageMedia{
			Name: media.Name,
			Size: media.Size,
			DC:   media.DC,
		}
	}

	m.Views = msg.Views
	m.Forwards = msg.Forwards

	return m
}
