package downloader

import (
	"context"
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/gabriel-vasile/mimetype"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/tg"
	"github.com/iyear/tdl/pkg/consts"
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

type Downloader struct {
	client   *tg.Client
	pw       progress.Writer
	partSize int
	threads  int
	iter     Iter
}

func New(client *tg.Client, partSize int, threads int, iter Iter) *Downloader {
	return &Downloader{
		client:   client,
		pw:       prog.New(),
		partSize: partSize,
		threads:  threads,
		iter:     iter,
	}
}

func (d *Downloader) Download(ctx context.Context, limit int) error {
	d.pw.Log(color.GreenString("All files will be downloaded to '%s' dir", consts.DownloadPath))

	d.pw.SetNumTrackersExpected(d.iter.Total(ctx))

	go d.pw.Render()

	wg, errctx := errgroup.WithContext(ctx)
	wg.SetLimit(limit)

	for d.iter.Next(ctx) {
		item, err := d.iter.Value(ctx)
		if err != nil {
			d.pw.Log(color.RedString("Get item failed: %v, skip...", err))
			continue
		}

		wg.Go(func() error {
			// d.pw.Log(color.MagentaString("name: %s,size: %s", item.Name, utils.Byte.FormatBinaryBytes(item.Size)))
			return d.download(errctx, item)
		})
	}

	err := wg.Wait()
	if err != nil {
		if errors.Is(err, context.Canceled) {
			d.pw.Log(color.RedString("Download aborted."))
		}

		d.pw.Stop()
		for d.pw.IsRenderInProgress() {
			time.Sleep(time.Millisecond * 10)
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

func (d *Downloader) download(ctx context.Context, item *Item) error {
	tracker := prog.AppendTracker(d.pw, item.Name, item.Size)
	filename := fmt.Sprintf("%s%s", utils.FS.GetNameWithoutExt(item.Name), TempExt)
	path := filepath.Join(consts.DownloadPath, filename)

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

	newfile := fmt.Sprintf("%s%s", strings.TrimSuffix(filename, TempExt), mime.Extension())
	if err = os.Rename(path, filepath.Join(consts.DownloadPath, newfile)); err != nil {
		return err
	}

	return nil
}
