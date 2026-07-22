package up

import (
	"fmt"

	"github.com/gotd/td/tg"

	"github.com/iyear/tdl/core/uploader"
)

type iterElem struct {
	task *uploadTask
}

func (e *iterElem) Media() []uploader.MediaDescriptor {
	return e.task.media
}

func (e *iterElem) To() tg.InputPeerClass {
	return e.task.to.InputPeer()
}

func (e *iterElem) Thread() int {
	return e.task.thread
}

func (e *iterElem) Remove() bool {
	return e.task.remove
}

func (e *iterElem) Paths() []string {
	return e.task.paths
}

func (e *iterElem) TotalSize() int64 {
	return e.task.totalSize
}

func (e *iterElem) Label() string {
	if e.task.label != "" {
		return e.task.label
	}
	return fmt.Sprintf("%d media -> %s(%d)", len(e.task.media), e.task.to.VisibleName(), e.task.to.ID())
}

func (e *iterElem) CaptionEntities(media uploader.MediaDescriptor) []tg.MessageEntityClass {
	return e.task.captionEntities(media)
}
