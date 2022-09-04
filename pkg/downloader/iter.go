package downloader

import (
	"context"
	"github.com/gotd/td/tg"
)

type Iter interface {
	Next(ctx context.Context) bool
	Value(ctx context.Context) (*Item, error)
	Total(ctx context.Context) int
}

type Item struct {
	InputFileLoc tg.InputFileLocationClass
	Name         string
	Size         int64
}
