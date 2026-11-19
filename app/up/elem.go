package up

import (
	"os"
	"path/filepath"

	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
	"github.com/samber/lo"

	"github.com/iyear/tdl/core/uploader"
)

type iterElem struct {
	file    *uploaderFile
	thumb   *uploaderFile
	to      peers.Peer
	caption *entity.Builder
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

func (e *iterElem) Caption() (string, []tg.MessageEntityClass) {
	return e.caption.Complete()
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

// groupElem is a group of files sent as one grouped message (album).
// All members share the same destination, and only the first member
// carries the caption, like official clients do.
type groupElem struct {
	elems []*iterElem
}

func (g *groupElem) Elems() []uploader.Elem {
	return lo.Map(g.elems, func(e *iterElem, _ int) uploader.Elem { return e })
}

func (g *groupElem) File() uploader.File {
	return g.elems[0].File()
}

func (g *groupElem) Thumb() (uploader.File, bool) {
	return g.elems[0].Thumb()
}

func (g *groupElem) Caption() (string, []tg.MessageEntityClass) {
	return g.elems[0].Caption()
}

func (g *groupElem) To() tg.InputPeerClass {
	return g.elems[0].To()
}

func (g *groupElem) Thread() int {
	return g.elems[0].Thread()
}

func (g *groupElem) AsPhoto() bool {
	return g.elems[0].AsPhoto()
}
