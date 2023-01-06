package downloader

import (
	"context"
	"errors"

	"github.com/gotd/td/tg"
)

var ErrSkip = errors.New("skip")

type Iter interface {
	Next(ctx context.Context) (*Item, error)
	Total(ctx context.Context) int
}

type Item struct {
	InputFileLoc tg.InputFileLocationClass
	Name         string
	Size         int64
	DC           int
	ChatID       int64
	MsgID        int
}
