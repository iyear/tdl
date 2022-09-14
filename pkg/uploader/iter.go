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

type Item struct {
	R    io.ReadCloser
	Name string
	MIME string
	Size int64
}
