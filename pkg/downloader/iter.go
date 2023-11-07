package downloader

import (
	"context"
	"errors"

	"github.com/iyear/tdl/pkg/tmedia"
)

var ErrSkip = errors.New("skip")

type Iter interface {
	Next(ctx context.Context) (*Item, error)
	Finish(ctx context.Context, id int) error
	Total(ctx context.Context) int
}

type Item struct {
	ID int // unique in iter
	*tmedia.Media
}
