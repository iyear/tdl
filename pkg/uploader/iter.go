package uploader

import (
	"context"
	"io"
)

type Iter interface {
	Next(ctx context.Context) bool
	Value(ctx context.Context) (*Item, error)
	Total(ctx context.Context) int
}

type ReadSeekCloser interface {
	io.Reader
	io.Seeker
	io.Closer
}

type Item struct {
	File  ReadSeekCloser
	Thumb ReadSeekCloser
	Name  string
	MIME  string
	Size  int64
}
