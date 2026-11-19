package uploader

import (
	"context"
	"io"

	"github.com/gotd/td/tg"
)

type Iter interface {
	Next(ctx context.Context) bool
	Value() Elem
	Err() error
}

type File interface {
	io.ReadSeeker
	Name() string
	Size() int64
}

type Elem interface {
	File() File
	Thumb() (File, bool)
	Caption() (string, []tg.MessageEntityClass)
	To() tg.InputPeerClass
	Thread() int
	AsPhoto() bool
}

// GroupElem is an Elem that contains multiple files sent as one grouped
// message (album). All member elems must share the same To() and Thread().
// Caption is only applied to the first member, like official clients do.
type GroupElem interface {
	Elem
	Elems() []Elem
}
