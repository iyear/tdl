package downloader

import (
	"context"
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/gabriel-vasile/mimetype"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/tg"
	"github.com/iyear/tdl/pkg/prog"
	"github.com/iyear/tdl/pkg/utils"
	"github.com/jedib0t/go-pretty/v6/progress"
	"golang.org/x/sync/errgroup"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const TempExt = ".tmp"

var formatter = utils.Byte.FormatBinaryBytes

type Downloader struct {
	client   *tg.Client
	pw       progress.Writer
	partSize int
	threads  int
	iter     Iter
	dir      string
}

func New(client *tg.Client, dir string, partSize int, threads int, iter Iter) *Downloader {
	return &Downloader{
		client:   client,
		pw:       prog.New(formatter),
		partSize: partSize,
		threads:  threads,
		iter:     iter,
		dir:      dir,
	}
}

func (d *Downloader) Download(ctx context.Context, limit int) error {
	d.pw.Log(color.GreenString("All files will be downloaded to '%s' dir", d.dir))

	total := d.iter.Total(ctx)
	d.pw.SetNumTrackersExpected(total)

	go d.pw.Render()

	wg, errctx := errgroup.WithContext(ctx)
	wg.SetLimit(limit)

	go runPS(errctx, d.pw)

	for i := 0; i < total; i++ {
		wg.Go(func() error {
			item, err := d.iter.Next(errctx)
			if err != nil {
				// skip error means we don't need to log error
				if !errors.Is(err, ErrSkip) {
					d.pw.Log(color.RedString("failed: %v", err))
				}
				return nil
			}
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

	for d.pw.IsRenderInProgress() {
		if d.pw.LengthActive() == 0 {
			d.pw.Stop()
		}
		time.Sleep(10 * time.Millisecond)
	}

	return nil
}

// safe filename for windows
var replacer = strings.NewReplacer(
	"/", "_", "\\", "_",
	":", "_", "*", "_",
	"?", "_", "\"", "_",
	"<", "_", ">", "_",
	"|", "_", " ", "_")

func (d *Downloader) download(ctx context.Context, item *Item) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	name := replacer.Replace(item.Name)
	tracker := prog.AppendTracker(d.pw, formatter, name, item.Size)
	filename := fmt.Sprintf("%s%s", name, TempExt)
	path := filepath.Join(d.dir, filename)

	f, err := os.Create(path)
	if err != nil {
		return err
	}

	_, err = downloader.NewDownloader().WithPartSize(d.partSize).
		Download(d.client, item.InputFileLoc).WithThreads(d.threads).
		Parallel(ctx, &writeAt{
			f:       f,
			tracker: tracker,
		})
	if err := f.Close(); err != nil {
		return err
	}
	if err != nil {
		return err
	}

	mime, err := mimetype.DetectFile(path)
	if err != nil {
		return err
	}

	// rename file, remove temp extension and add real extension
	newfile := strings.TrimSuffix(filename, TempExt)
	ext := mime.Extension()
	if ext != "" && (filepath.Ext(newfile) != ext) {
		newfile = utils.FS.GetNameWithoutExt(newfile) + ext
	}

	if err = os.Rename(path, filepath.Join(d.dir, newfile)); err != nil {
		return err
	}

	return nil
}
