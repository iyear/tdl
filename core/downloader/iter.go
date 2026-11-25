package downloader

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

type Elem interface {
	File() File
	To() io.WriterAt
	// Offset is the number of bytes already downloaded. A non-zero offset
	// means the file is partially downloaded and should be resumed.
	Offset() int64

	AsTakeout() bool
}

type File interface {
	Location() tg.InputFileLocationClass
	Size() int64
	DC() int
}
