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
				// Check for network errors first (before context.Canceled check)
				// Network errors may have context.Canceled wrapped in them, but they're recoverable
				var netErr *NetworkError
				var stallErr *ServerStallingError
				if errors.As(err, &netErr) || errors.As(err, &stallErr) {
					// These are non-fatal, progress handler will display them
					// Don't return error to errgroup, just pass to OnDone for logging
					return nil
				}

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
		// Check for network EOF errors (connection dropped during transfer)
		// These should not be fatal - file will be retried with --continue flag
		if isNetworkError(err) {
			return &NetworkError{underlying: err}
		}
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

// NetworkError indicates a network/connection error during download (non-fatal, can retry)
type NetworkError struct {
	underlying error
}

func (e *NetworkError) Error() string {
	return "network connection error during download"
}

func (e *NetworkError) Unwrap() error {
	return e.underlying
}

// Known error patterns from gotd library that indicate server stalling/refusing connection
var stallingErrorPatterns = [][]string{
	{"create invoker", "context canceled"},
	{"export auth", "context canceled"},
	{"transfer", "context canceled"},
}

// isServerStallingError checks if the error indicates server is stalling the download
// This matches error patterns from the gotd/td library when the server refuses
// to establish a connection, similar to how the retry middleware handles gotd errors
func isServerStallingError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Check all known stalling patterns
	for _, pattern := range stallingErrorPatterns {
		allMatch := true
		for _, substring := range pattern {
			if !strings.Contains(errStr, substring) {
				allMatch = false
				break
			}
		}
		if allMatch {
			return true
		}
	}

	return false
}

// Known network error patterns that indicate connection issues (non-fatal)
var networkErrorPatterns = [][]string{
	{"EOF"},
	{"connection reset"},
	{"broken pipe"},
	{"i/o timeout"},
}

// isNetworkError checks if the error is a network/connection error
func isNetworkError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Check all known network error patterns
	for _, pattern := range networkErrorPatterns {
		allMatch := true
		for _, substring := range pattern {
			if !strings.Contains(errStr, substring) {
				allMatch = false
				break
			}
		}
		if allMatch {
			return true
		}
	}

	return false
}
