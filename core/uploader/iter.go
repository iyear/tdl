package uploader

import (
	"context"
	"io"

	"github.com/gotd/td/telegram/message"
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
	Caption() []message.StyledTextOption
	To() tg.InputPeerClass
	Thread() int
	AsPhoto() bool
}
