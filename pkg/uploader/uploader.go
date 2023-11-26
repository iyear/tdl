package uploader

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
	"go.uber.org/multierr"
	"golang.org/x/sync/errgroup"

	"github.com/iyear/tdl/pkg/utils"
)

type Uploader struct {
	opts Options
}

type Options struct {
	Client   *tg.Client
	PartSize int
	Threads  int
	Iter     Iter
	Progress Progress
}

func New(o Options) *Uploader {
	return &Uploader{opts: o}
}

func (u *Uploader) Upload(ctx context.Context, limit int) error {
	wg, wgctx := errgroup.WithContext(ctx)
	wg.SetLimit(limit)

	for u.opts.Iter.Next(wgctx) {
		elem := u.opts.Iter.Value()

		wg.Go(func() (rerr error) {
			u.opts.Progress.OnAdd(elem)
			defer u.opts.Progress.OnDone(elem, rerr)

			if err := u.upload(wgctx, elem); err != nil {
				// canceled by user, so we directly return error to stop all
				if errors.Is(err, context.Canceled) {
					return errors.Wrap(err, "upload")
				}

				// don't return error, just log it
			}

			return nil
		})
	}

	if err := u.opts.Iter.Err(); err != nil {
		return errors.Wrap(err, "iter")
	}

	return wg.Wait()
}

func (u *Uploader) upload(ctx context.Context, elem *Elem) (rerr error) {
	defer func() {
		if rerr == nil && elem.Remove {
			multierr.AppendInto(&rerr, elem.File.Remove())
			multierr.AppendInto(&rerr, elem.Thumb.Remove())
		}
	}()

	defer multierr.AppendInvoke(&rerr, multierr.Close(elem.File))
	defer multierr.AppendInvoke(&rerr, multierr.Close(elem.Thumb))

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	up := uploader.NewUploader(u.opts.Client).
		WithPartSize(u.opts.PartSize).
		WithThreads(u.opts.Threads).
		WithProgress(&wrapProcess{
			elem:    elem,
			process: u.opts.Progress,
		})

	f, err := up.Upload(ctx, uploader.NewUpload(elem.Name, elem.File, elem.Size))
	if err != nil {
		return errors.Wrap(err, "upload file")
	}

	caption := []message.StyledTextOption{
		styling.Code(elem.Name),
		styling.Plain(" - "),
		styling.Code(elem.MIME),
	}
	doc := message.UploadedDocument(f, caption...).MIME(elem.MIME).Filename(elem.Name)
	// upload thumbnail TODO(iyear): maybe still unavailable
	if thumb, err := uploader.NewUploader(u.opts.Client).
		FromReader(ctx, fmt.Sprintf("%s.thumb", elem.Name), elem.Thumb); err == nil {
		doc = doc.Thumb(thumb)
	}

	var media message.MediaOption = doc

	switch {
	case utils.Media.IsImage(elem.MIME) && elem.Photo:
		// upload as photo
		media = message.UploadedPhoto(f, caption...)
	case utils.Media.IsVideo(elem.MIME):
		// reset reader
		if _, err = elem.File.Seek(0, io.SeekStart); err != nil {
			return errors.Wrap(err, "seek file")
		}
		if dur, w, h, err := utils.Media.GetMP4Info(elem.File); err == nil {
			// #132. There may be some errors, but we can still upload the file
			media = doc.Video().
				Duration(time.Duration(dur)*time.Second).
				Resolution(w, h).
				SupportsStreaming()
		}
	case utils.Media.IsAudio(elem.MIME):
		media = doc.Audio().Title(utils.FS.GetNameWithoutExt(elem.Name))
	}

	_, err = message.NewSender(u.opts.Client).
		WithUploader(up).
		To(elem.To.InputPeer()).
		Media(ctx, media)
	if err != nil {
		return errors.Wrap(err, "send message")
	}

	return nil
}
