package up

import (
	"context"
	"io"
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
	files []*file
	to    peers.Peer
	photo bool

	cur  int
	err  error
	file *uploader.Elem
}

func newIter(files []*file, to peers.Peer, photo bool) *iter {
	return &iter{
		files: files,
		cur:   0,
		err:   nil,
		file:  nil,
		to:    to,
		photo: photo,
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

	var thumb io.ReadSeekCloser = nopReadSeekCloser{}
	// has thumbnail
	if cur.thumb != "" {
		tMime, err := mimetype.DetectFile(cur.thumb)
		if err != nil || !utils.Media.IsImage(tMime.String()) { // TODO(iyear): jpg only
			i.err = errors.Wrapf(err, "invalid thumbnail file: %v", cur.thumb)
			return false
		}
		thumb, err = os.Open(cur.thumb)
		if err != nil {
			i.err = errors.Wrap(err, "open thumbnail file")
			return false
		}
	}

	i.file = &uploader.Elem{
		File:  f,
		Thumb: thumb,
		Name:  filepath.Base(f.Name()),
		MIME:  fMime.String(),
		Size:  stat.Size(),
		To:    i.to,
		Photo: i.photo,
	}

	return true
}

func (i *iter) Value() *uploader.Elem {
	return i.file
}

func (i *iter) Err() error {
	return i.err
}

type nopReadSeekCloser struct{}

func (nopReadSeekCloser) Read(_ []byte) (n int, err error) {
	return 0, errors.New("nopReadSeekCloser")
}

func (nopReadSeekCloser) Seek(_ int64, _ int) (int64, error) {
	return 0, errors.New("nopReadSeekCloser")
}

func (nopReadSeekCloser) Close() error {
	return nil
}
