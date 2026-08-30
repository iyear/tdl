package thumbnail

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io"
)

const maxDimension = 320

// File is an in-memory thumbnail ready to be uploaded.
type File interface {
	io.ReadSeeker
	Name() string
	Size() int64
}

type memoryFile struct {
	*bytes.Reader
	name string
}

func (f *memoryFile) Name() string { return f.name }

func (f *memoryFile) Size() int64 { return int64(f.Reader.Size()) }

// NewBlack creates an opaque black JPEG in memory. It preserves the video's
// aspect ratio, fits within Telegram's 320x320 thumbnail limit, and never
// touches the filesystem.
func NewBlack(width, height int) (File, error) {
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid video resolution %dx%d", width, height)
	}

	scale := float64(maxDimension) / float64(max(width, height))
	width = max(1, int(float64(width)*scale))
	height = max(1, int(float64(height)*scale))

	canvas := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(canvas, canvas.Bounds(), image.NewUniform(color.Black), image.Point{}, draw.Src)

	var encoded bytes.Buffer
	if err := jpeg.Encode(&encoded, canvas, &jpeg.Options{Quality: 90}); err != nil {
		return nil, fmt.Errorf("encode black thumbnail: %w", err)
	}
	return &memoryFile{
		Reader: bytes.NewReader(encoded.Bytes()),
		name:   "video-thumbnail.jpg",
	}, nil
}
