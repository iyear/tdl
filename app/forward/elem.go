package forward

import (
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"

	"github.com/iyear/tdl/core/forwarder"
)

// RenameFunc is a function that computes a renamed filename for a given message.
// It takes the source peer and message, returning the new filename.
// Returns empty string to keep the original filename.
type RenameFunc func(from peers.Peer, msg *tg.Message) string

type iterElem struct {
	from         peers.Peer
	msg          *tg.Message
	to           peers.Peer
	thread       int
	modeOverride forwarder.Mode
	renameFunc   RenameFunc // closure to compute filename per message
	opts         iterOptions
}

func (i *iterElem) Mode() forwarder.Mode {
	if i.modeOverride.IsValid() {
		return i.modeOverride
	}
	return i.opts.mode
}

func (i *iterElem) From() peers.Peer { return i.from }

func (i *iterElem) Msg() *tg.Message { return i.msg }

func (i *iterElem) To() peers.Peer { return i.to }

func (i *iterElem) Thread() int { return i.thread }

func (i *iterElem) AsSilent() bool { return i.opts.silent }

func (i *iterElem) AsDryRun() bool { return i.opts.dryRun }

func (i *iterElem) AsGrouped() bool { return i.opts.grouped }

// ComputeRenamedFilename computes the renamed filename for a given message.
// This allows each message in an album to have its own unique filename.
func (i *iterElem) ComputeRenamedFilename(msg *tg.Message) string {
	if i.renameFunc == nil {
		return ""
	}
	return i.renameFunc(i.from, msg)
}
