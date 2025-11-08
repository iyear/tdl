package downloader

import (
	"context"
	"strings"

	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram/downloader"
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

	_, err := downloader.NewDownloader().WithPartSize(MaxPartSize).
		Download(client, elem.File().Location()).
		WithThreads(tutil.BestThreads(elem.File().Size(), d.opts.Threads)).
		Parallel(ctx, newWriteAt(elem, d.opts.Progress, MaxPartSize))
	if err != nil {
		// Check if this is a "create invoker" error with "context canceled" which indicates
		// the server is stalling or refusing the connection (hangs at 0%)
		if isServerStallingError(err) {
			return &ServerStallingError{underlying: err}
		}
		return errors.Wrap(err, "download")
	}

	return nil
}

// ServerStallingError indicates the server is stalling/refusing the download
type ServerStallingError struct {
	underlying error
}

func (e *ServerStallingError) Error() string {
	return "server stalling or refusing download"
}

func (e *ServerStallingError) Unwrap() error {
	return e.underlying
}

// isServerStallingError checks if the error indicates server is stalling the download
func isServerStallingError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Check for the specific "create invoker" + "context canceled" pattern
	// This occurs when the server refuses to establish the connection for download
	if strings.Contains(errStr, "create invoker") && strings.Contains(errStr, "context canceled") {
		return true
	}

	// Also check for "export auth" + "context canceled" pattern
	if strings.Contains(errStr, "export auth") && strings.Contains(errStr, "context canceled") {
		return true
	}

	return false
}
