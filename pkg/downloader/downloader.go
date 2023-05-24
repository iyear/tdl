package downloader

import (
	"context"
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/gabriel-vasile/mimetype"
	"github.com/gotd/td/telegram/downloader"
	"github.com/iyear/tdl/pkg/dcpool"
	"github.com/iyear/tdl/pkg/logger"
	"github.com/iyear/tdl/pkg/prog"
	"github.com/iyear/tdl/pkg/utils"
	"github.com/jedib0t/go-pretty/v6/progress"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const TempExt = ".tmp"

var formatter = utils.Byte.FormatBinaryBytes

type Downloader struct {
	pool                 dcpool.Pool
	pw                   progress.Writer
	partSize             int
	threads              int
	iter                 Iter
	dir                  string
	rewriteExt, skipSame bool
	takeout              bool
}

type Options struct {
	Pool       dcpool.Pool
	Dir        string
	RewriteExt bool
	SkipSame   bool
	PartSize   int
	Threads    int
	Iter       Iter
	Takeout    bool
}

func New(opts *Options) *Downloader {
	return &Downloader{
		pool:       opts.Pool,
		pw:         prog.New(formatter),
		partSize:   opts.PartSize,
		threads:    opts.Threads,
		iter:       opts.Iter,
		dir:        opts.Dir,
		rewriteExt: opts.RewriteExt,
		skipSame:   opts.SkipSame,
		takeout:    opts.Takeout,
	}
}

func (d *Downloader) Download(ctx context.Context, limit int) error {
	color.Green("All files will be downloaded to '%s' dir", d.dir)

	total := d.iter.Total(ctx)
	d.pw.SetNumTrackersExpected(total)

	go d.renderPinned(ctx, d.pw)
	go d.pw.Render()

	wg, errctx := errgroup.WithContext(ctx)
	wg.SetLimit(limit)

	for i := 0; i < total; i++ {
		item, err := d.iter.Next(errctx)
		if err != nil {
			logger.From(errctx).Debug("Iter next failed",
				zap.Int("index", i), zap.String("error", err.Error()))
			// skip error means we don't need to log error
			if !errors.Is(err, ErrSkip) && !errors.Is(err, context.Canceled) {
				d.pw.Log(color.RedString("failed: %v", err))
			}
			continue
		}

		wg.Go(func() error {
			return d.download(errctx, item)
		})
	}

	err := wg.Wait()
	if err != nil {
		d.pw.Stop()
		for d.pw.IsRenderInProgress() {
			time.Sleep(time.Millisecond * 10)
		}

		if errors.Is(err, context.Canceled) {
			color.Red("Download aborted.")
		}
		return err
	}

	prog.Wait(d.pw)

	return nil
}

func (d *Downloader) download(ctx context.Context, item *Item) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	logger.From(ctx).Debug("Start download item",
		zap.Any("item", item))

	if d.skipSame {
		if stat, err := os.Stat(filepath.Join(d.dir, item.Name)); err == nil {
			if utils.FS.GetNameWithoutExt(item.Name) == utils.FS.GetNameWithoutExt(stat.Name()) &&
				stat.Size() == item.Size {
				return nil
			}
		}
	}
	tracker := prog.AppendTracker(d.pw, formatter, item.Name, item.Size)
	filename := fmt.Sprintf("%s%s", item.Name, TempExt)
	path := filepath.Join(d.dir, filename)

	// #113. If path contains dirs, create it. So now we support nested dirs.
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}

	client := d.pool.Client(item.DC)
	if d.takeout {
		client = d.pool.Takeout(item.DC)
	}

	_, err = downloader.NewDownloader().WithPartSize(d.partSize).
		Download(client, item.InputFileLoc).
		WithThreads(d.bestThreads(item.Size)).
		Parallel(ctx, newWriteAt(f, tracker, d.partSize))
	if err := f.Close(); err != nil {
		return err
	}
	if err != nil {
		return err
	}

	// rename file, remove temp extension and add real extension
	newfile := strings.TrimSuffix(filename, TempExt)

	if d.rewriteExt {
		mime, err := mimetype.DetectFile(path)
		if err != nil {
			return err
		}
		ext := mime.Extension()
		if ext != "" && (filepath.Ext(newfile) != ext) {
			newfile = utils.FS.GetNameWithoutExt(newfile) + ext
		}
	}

	if err = os.Rename(path, filepath.Join(d.dir, newfile)); err != nil {
		return err
	}

	return d.iter.Finish(ctx, item.ID)
}

// threads level
// TODO(iyear): more practice to find best number
var threads = []struct {
	threads int
	size    int64
}{
	{1, 1 << 20},
	{2, 5 << 20},
	{4, 20 << 20},
	{8, 50 << 20},
}

// Get best threads num for download, based on file size
func (d *Downloader) bestThreads(size int64) int {
	for _, t := range threads {
		if size < t.size {
			return min(t.threads, d.threads)
		}
	}
	return d.threads
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
