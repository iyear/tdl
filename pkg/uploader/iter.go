package uploader

import (
	"context"
	"io"

	"github.com/gotd/td/telegram/peers"
)

type Iter interface {
	Next(ctx context.Context) bool
	Value() *Elem
	Err() error
}

type Elem struct {
	File  io.ReadSeekCloser
	Thumb io.ReadSeekCloser
	Name  string
	MIME  string
	Size  int64
	To    peers.Peer
	Photo bool
}
