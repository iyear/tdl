package forwarder

import (
	"context"

	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
)

type Iter interface {
	Next(ctx context.Context) bool
	Value() Elem
	Err() error
}

type Elem interface {
	Mode() Mode

	From() peers.Peer
	Msg() *tg.Message
	To() peers.Peer
	Thread() int // reply to message/topic

	AsSilent() bool
	AsDryRun() bool
	AsGrouped() bool // detect and forward grouped messages
	// ComputeRenamedFilename computes the renamed filename for a given message.
	// This allows each message in an album to have its own unique filename based on its ID.
	// Returns empty string to keep the original filename.
	ComputeRenamedFilename(msg *tg.Message) string
}
