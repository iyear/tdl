package chat

import (
	"github.com/gotd/td/tg"

	"github.com/iyear/tdl/pkg/tmedia"
	"github.com/iyear/tdl/pkg/utils"
)

type message struct {
	Mentioned     bool         `comment:"Whether we were mentioned in this message"`
	Silent        bool         `comment:"Whether this is a silent message (no notification triggered)"`
	FromScheduled bool         `comment:"Whether this is a scheduled message"`
	Pinned        bool         `comment:"Whether this message is pinned"`
	ID            int          `comment:"ID of the message"`
	FromID        int64        `comment:"ID of the sender of the message"`
	Date          int          `comment:"Date of the message"`
	Message       string       `comment:"The message"`
	Media         messageMedia `comment:"Media attachment"`
	Views         int          `comment:"View count"`
	Forwards      int          `comment:"Forward count"`
}

type messageMedia struct {
	Name string `comment:"File name"`
	Size int64  `comment:"File size. Unit: Byte"`
	DC   int    `comment:"DC ID"`
}

func covertMessage(msg *tg.Message) (m *message) {
	m = &message{}

	m.Mentioned = msg.Mentioned
	m.Silent = msg.Silent
	m.FromScheduled = msg.FromScheduled
	m.Pinned = msg.Pinned
	m.ID = msg.ID
	m.FromID = utils.Telegram.GetPeerID(msg.FromID)
	m.Date = msg.Date
	m.Message = msg.Message

	if media, ok := tmedia.GetMedia(msg); ok {
		m.Media = messageMedia{
			Name: media.Name,
			Size: media.Size,
			DC:   media.DC,
		}
	}

	m.Views = msg.Views
	m.Forwards = msg.Forwards

	return m
}
