package forwarder

import (
	"context"
	"io"
	"os"

	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
	"go.uber.org/atomic"
	"go.uber.org/multierr"

	"github.com/iyear/tdl/pkg/tmedia"
	"github.com/iyear/tdl/pkg/utils"
)

type cloneOptions struct {
	elem     Elem
	media    *tmedia.Media
	progress progressAdd
}

type progressAdd interface {
	add(n int64)
}

func (f *Forwarder) cloneMedia(ctx context.Context, opts cloneOptions, dryRun bool) (_ tg.InputFileClass, rerr error) {
	// if dry run, just return empty input file
	if dryRun {
		// directly call progress callback
		opts.progress.add(opts.media.Size * 2)

		return &tg.InputFile{}, nil
	}

	temp, err := os.CreateTemp("", "tdl_*")
	if err != nil {
		return nil, errors.Wrap(err, "create temp file")
	}
	defer func() {
		multierr.AppendInto(&rerr, temp.Close())
		multierr.AppendInto(&rerr, os.Remove(temp.Name()))
	}()

	threads := utils.Telegram.BestThreads(opts.media.Size, f.opts.Threads)

	_, err = downloader.NewDownloader().
		WithPartSize(f.opts.PartSize).
		Download(f.opts.Pool.Client(ctx, opts.media.DC), opts.media.InputFileLoc).
		WithThreads(threads).
		Parallel(ctx, writeAt{
			f:    temp,
			opts: opts,
		})
	if err != nil {
		return nil, errors.Wrap(err, "download")
	}

	var file tg.InputFileClass

	if _, err = temp.Seek(0, io.SeekStart); err != nil {
		return nil, errors.Wrap(err, "seek")
	}

	upload := uploader.NewUpload(opts.media.Name, temp, opts.media.Size)
	file, err = uploader.NewUploader(f.opts.Pool.Default(ctx)).
		WithPartSize(f.opts.PartSize).
		WithThreads(threads).
		WithProgress(uploaded{
			opts: opts,
			prev: atomic.NewInt64(0),
		}).
		Upload(ctx, upload)
	if err != nil {
		return nil, errors.Wrap(err, "upload")
	}

	return file, nil
}

type writeAt struct {
	f    io.WriterAt
	opts cloneOptions
}

func (w writeAt) WriteAt(p []byte, off int64) (int, error) {
	n, err := w.f.WriteAt(p, off)
	if err != nil {
		return 0, err
	}

	w.opts.progress.add(int64(n))

	return n, nil
}

type uploaded struct {
	opts cloneOptions
	prev *atomic.Int64
}

func (u uploaded) Chunk(_ context.Context, state uploader.ProgressState) error {
	u.opts.progress.add(state.Uploaded - u.prev.Swap(state.Uploaded))

	return nil
}
