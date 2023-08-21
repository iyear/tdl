package up

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gabriel-vasile/mimetype"
	"go.uber.org/zap"

	"github.com/iyear/tdl/pkg/logger"
	"github.com/iyear/tdl/pkg/uploader"
	"github.com/iyear/tdl/pkg/utils"
)

type file struct {
	file  string
	thumb string
}

type iter struct {
	files  []*file
	cur    int
	remove bool
}

func newIter(files []*file, remove bool) *iter {
	return &iter{
		files:  files,
		cur:    -1,
		remove: remove,
	}
}

func (i *iter) Next(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	default:
	}

	i.cur++

	return i.cur != len(i.files)
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
		ID:    i.cur,
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

func (i *iter) Finish(ctx context.Context, id int) {
	if !i.remove {
		return
	}

	l := logger.From(ctx)
	if err := os.Remove(i.files[id].file); err != nil {
		l.Error("remove file failed", zap.Error(err))
		return
	}
	l.Info("remove file success", zap.String("file", i.files[id].file))
}
