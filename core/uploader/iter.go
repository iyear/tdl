package uploader

import (
	"context"

	"github.com/gotd/td/tg"
)

type Iter interface {
	Next(ctx context.Context) bool
	Value() Elem
	Err() error
}

type MediaType string

const (
	MediaTypePhoto    MediaType = "photo"
	MediaTypeVideo    MediaType = "video"
	MediaTypeDocument MediaType = "document"
)

type MediaDescriptor struct {
	Path      string
	Thumb     string
	Width     int
	Height    int
	Duration  int
	Mime      string
	Streaming bool
	Caption   string
	AlbumID   string
	Type      MediaType
}

type Elem interface {
	Media() []MediaDescriptor
	To() tg.InputPeerClass
	Thread() int
	Remove() bool
	Paths() []string
	TotalSize() int64
	Label() string
	CaptionEntities(media MediaDescriptor) []tg.MessageEntityClass
}
