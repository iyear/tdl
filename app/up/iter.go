package up

import (
	"context"
	"fmt"
	"github.com/gabriel-vasile/mimetype"
	"github.com/iyear/tdl/pkg/uploader"
	"github.com/iyear/tdl/pkg/utils"
	"os"
	"path/filepath"
)

type file struct {
	file  string
	thumb string
}

type iter struct {
	files []*file
	cur   int
}

func newIter(files []*file) *iter {
	return &iter{
		files: files,
		cur:   -1,
	}
}

func (i *iter) Next(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	default:
	}

	i.cur++

	if i.cur == len(i.files) {
		return false
	}

	return true
}

func (i *iter) Value(ctx context.Context) (*uploader.Item, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	cur := i.files[i.cur]

	fMime, err := mimetype.DetectFile(cur.file)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(cur.file)
	if err != nil {
		return nil, err
	}

	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}

	var thumb *os.File
	// has thumbnail
	if cur.thumb != "" {
		tMime, err := mimetype.DetectFile(cur.thumb)
		if err != nil || !utils.Media.IsImage(tMime.String()) { // TODO(iyear): jpg only
			return nil, fmt.Errorf("invalid thumbnail file: %s", cur.thumb)
		}
		thumb, err = os.Open(cur.thumb)
		if err != nil {
			return nil, err
		}
	}

	return &uploader.Item{
		File:  f,
		Thumb: thumb,
		Name:  filepath.Base(f.Name()),
		MIME:  fMime.String(),
		Size:  stat.Size(),
	}, nil
}

func (i *iter) Total(_ context.Context) int {
	return len(i.files)
}
