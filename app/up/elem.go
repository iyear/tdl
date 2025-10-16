package up

import (
	"os"
	"path/filepath"

	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"

	"github.com/iyear/tdl/core/uploader"
)

type iterElem struct {
	file    *uploaderFile
	thumb   *uploaderFile
	to      peers.Peer
	caption []message.StyledTextOption
	thread  int

	asPhoto bool
	remove  bool
}

func (e *iterElem) File() uploader.File {
	return e.file
}

func (e *iterElem) Thumb() (uploader.File, bool) {
	if e.thumb == nil {
		return nil, false
	}
	return e.thumb, true
}

func (e *iterElem) Caption() []message.StyledTextOption {
	return e.caption
}

func (e *iterElem) To() tg.InputPeerClass {
	return e.to.InputPeer()
}

func (e *iterElem) Thread() int {
	return e.thread
}

func (e *iterElem) AsPhoto() bool {
	return e.asPhoto
}

type uploaderFile struct {
	*os.File
	size int64
}

func (u *uploaderFile) Name() string {
	return filepath.Base(u.File.Name())
}

func (u *uploaderFile) Size() int64 {
	return u.size
}
