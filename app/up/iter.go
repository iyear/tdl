package up

import (
	"context"
	"os"
	"path/filepath"

	"github.com/gabriel-vasile/mimetype"
	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram/peers"

	"github.com/iyear/tdl/pkg/uploader"
	"github.com/iyear/tdl/pkg/utils"
)

type file struct {
	file  string
	thumb string
}

type iter struct {
	files  []*file
	to     peers.Peer
	photo  bool
	remove bool

	cur  int
	err  error
	file *uploader.Elem
}

func newIter(files []*file, to peers.Peer, photo, remove bool) *iter {
	return &iter{
		files:  files,
		cur:    0,
		err:    nil,
		file:   nil,
		to:     to,
		photo:  photo,
		remove: remove,
	}
}

func (i *iter) Next(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		i.err = ctx.Err()
		return false
	default:
	}

	if i.cur >= len(i.files) || i.err != nil {
		return false
	}

	cur := i.files[i.cur]
	i.cur++

	fMime, err := mimetype.DetectFile(cur.file)
	if err != nil {
		i.err = errors.Wrap(err, "detect mime")
		return false
	}

	f, err := os.Open(cur.file)
	if err != nil {
		i.err = errors.Wrap(err, "open file")
		return false
	}

	stat, err := f.Stat()
	if err != nil {
		i.err = errors.Wrap(err, "stat file")
		return false
	}

	var thumb uploader.File = nopFile{}
	// has thumbnail
	if cur.thumb != "" {
		tMime, err := mimetype.DetectFile(cur.thumb)
		if err != nil || !utils.Media.IsImage(tMime.String()) { // TODO(iyear): jpg only
			i.err = errors.Wrapf(err, "invalid thumbnail file: %v", cur.thumb)
			return false
		}
		thumbFile, err := os.Open(cur.thumb)
		if err != nil {
			i.err = errors.Wrap(err, "open thumbnail file")
			return false
		}

		thumb = uploaderFile{thumbFile}
	}

	i.file = &uploader.Elem{
		File:   uploaderFile{f},
		Thumb:  thumb,
		Name:   filepath.Base(f.Name()),
		MIME:   fMime.String(),
		Size:   stat.Size(),
		To:     i.to,
		Photo:  i.photo,
		Remove: i.remove,
	}

	return true
}

func (i *iter) Value() *uploader.Elem {
	return i.file
}

func (i *iter) Err() error {
	return i.err
}

type nopFile struct{}

func (nopFile) Read(_ []byte) (n int, err error) {
	return 0, errors.New("nopFile")
}

func (nopFile) Seek(_ int64, _ int) (int64, error) {
	return 0, errors.New("nopFile")
}

func (nopFile) Close() error {
	return nil
}

func (nopFile) Remove() error {
	return nil
}

type uploaderFile struct {
	*os.File
}

func (u uploaderFile) Remove() error {
	return os.Remove(u.Name())
}
