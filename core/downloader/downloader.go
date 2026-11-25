package downloader

import (
	"context"
	"io"

	"github.com/go-faster/errors"
	"github.com/gotd/contrib/partio"
	"github.com/gotd/contrib/tg_io"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/tg"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/iyear/tdl/core/dcpool"
	"github.com/iyear/tdl/core/logctx"
	"github.com/iyear/tdl/core/util/tutil"
)

// MaxPartSize refer to https://core.telegram.org/api/files#downloading-files
const MaxPartSize = 1024 * 1024

type Downloader struct {
	opts Options
}

type Options struct {
	Pool     dcpool.Pool
	Threads  int
	Iter     Iter
	Progress Progress
}

func New(opts Options) *Downloader {
	return &Downloader{
		opts: opts,
	}
}

func (d *Downloader) Download(ctx context.Context, limit int) error {
	wg, wgctx := errgroup.WithContext(ctx)
	wg.SetLimit(limit)

	for d.opts.Iter.Next(wgctx) {
		elem := d.opts.Iter.Value()

		wg.Go(func() (rerr error) {
			d.opts.Progress.OnAdd(elem)
			defer func() { d.opts.Progress.OnDone(elem, rerr) }()

			if err := d.download(wgctx, elem); err != nil {
				// canceled by user, so we directly return error to stop all
				if errors.Is(err, context.Canceled) {
					return errors.Wrap(err, "download")
				}

				// don't return error, just log it
				logctx.
					From(ctx).
					Error("Download error",
						zap.Any("element", elem),
						zap.Error(err),
					)
			}

			return nil
		})
	}

	if err := d.opts.Iter.Err(); err != nil {
		return errors.Wrap(err, "iter")
	}

	return wg.Wait()
}

func (d *Downloader) download(ctx context.Context, elem Elem) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	logctx.From(ctx).Debug("Start download elem",
		zap.Any("elem", elem))

	client := d.opts.Pool.Client(ctx, elem.File().DC())
	if elem.AsTakeout() {
		client = d.opts.Pool.Takeout(ctx, elem.File().DC())
	}

	// gotd's parallel downloader always starts at offset 0, so a partially
	// downloaded file would be re-fetched from the beginning. Resume it by
	// streaming the remainder from the saved offset instead.
	if offset := elem.Offset(); offset > 0 {
		return d.resume(ctx, client, elem, offset)
	}

	_, err := downloader.NewDownloader().WithPartSize(MaxPartSize).
		Download(client, elem.File().Location()).
		WithThreads(tutil.BestThreads(elem.File().Size(), d.opts.Threads)).
		Parallel(ctx, newWriteAt(elem, d.opts.Progress, MaxPartSize))
	if err != nil {
		return errors.Wrap(err, "download")
	}

	return nil
}

// resume continues an interrupted download from the given offset, writing the
// remainder of the file to the already partially written output.
func (d *Downloader) resume(ctx context.Context, client *tg.Client, elem Elem, offset int64) error {
	to, ok := elem.To().(io.WriteSeeker)
	if !ok {
		return errors.New("output does not support seeking")
	}
	if _, err := to.Seek(offset, io.SeekStart); err != nil {
		return errors.Wrap(err, "seek output")
	}

	streamer := partio.NewStreamer(
		tg_io.NewDownloader(client).ChunkSource(elem.File().Size(), elem.File().Location()),
		MaxPartSize,
	)

	if err := streamer.StreamAt(ctx, offset, newCountWriter(to, elem, d.opts.Progress, offset)); err != nil {
		return errors.Wrap(err, "download")
	}

	return nil
}
