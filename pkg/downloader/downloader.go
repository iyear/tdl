package downloader

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/gabriel-vasile/mimetype"
	"github.com/gotd/td/telegram/downloader"
	"github.com/iyear/tdl/pkg/dcpool"
	"github.com/iyear/tdl/pkg/prog"
	"github.com/iyear/tdl/pkg/utils"
	"github.com/jedib0t/go-pretty/v6/progress"
	"golang.org/x/sync/errgroup"
)

const TempExt = ".tmp"
const ResExt = ".tgresume"

var formatter = utils.Byte.FormatBinaryBytes

type Downloader struct {
	pool                 dcpool.Pool
	pw                   progress.Writer
	partSize             int
	threads              int
	iter                 Iter
	dir                  string
	rewriteExt, skipSame bool
}

func New(pool dcpool.Pool, dir string, rewriteExt, skipSame bool, partSize int, threads int, iter Iter) *Downloader {
	return &Downloader{
		pool:       pool,
		pw:         prog.New(formatter),
		partSize:   partSize,
		threads:    threads,
		iter:       iter,
		dir:        dir,
		rewriteExt: rewriteExt,
		skipSame:   skipSame,
	}
}

func (d *Downloader) Download(ctx context.Context, limit int) error {
	d.pw.Log(color.GreenString("All files will be downloaded to '%s' dir", d.dir))

	total := d.iter.Total(ctx)
	d.pw.SetNumTrackersExpected(total)

	go d.renderPinned(ctx, d.pw)
	go d.pw.Render()

	wg, errctx := errgroup.WithContext(ctx)
	wg.SetLimit(limit)

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

	name := replacer.Replace(item.Name) // tmp file finished writing here and is bein renamed to templayedname
	if d.skipSame {
		if stat, err := os.Stat(filepath.Join(d.dir, name)); err == nil {
			if utils.FS.GetNameWithoutExt(name) == utils.FS.GetNameWithoutExt(stat.Name()) &&
				stat.Size() == item.Size {
				return nil
			}
		}
	}
	tracker := prog.AppendTracker(d.pw, formatter, name, item.Size)
	filename := fmt.Sprintf("%s%s", name, TempExt)
	path := filepath.Join(d.dir, filename)

	f, err := os.Create(path)
	if err != nil {
		return err
	}

	_, err = downloader.NewDownloader().WithPartSize(d.partSize).
		Download(d.pool.Client(item.DC), item.InputFileLoc).WithThreads(d.threads).
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

	// open tgresumefile
	// Write completed MessageID to the tgresume file.
	resfile := fmt.Sprintf("%d%s", item.ChatID, ResExt)

	f2, err := os.OpenFile(resfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModeAppend|0644)
	if err != nil {
		return err
	}

	f2.WriteString(fmt.Sprintf("%d\n", item.MsgID))
	f2.Sync()
	f2.Close()

	return nil
}
