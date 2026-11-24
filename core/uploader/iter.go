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

// GroupElem is an Elem whose members form one grouped upload task.
// A single member is sent on its own. Multiple members must share the same
// To() and Thread() and have compatible media types.
type GroupElem interface {
	Elem
	Elems() []Elem
}
