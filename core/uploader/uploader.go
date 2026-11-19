package uploader

import (
	"context"
	"io"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"

	"github.com/iyear/tdl/core/util/fsutil"
	"github.com/iyear/tdl/core/util/mediautil"
)

// MaxPartSize refer to https://core.telegram.org/api/files#uploading-files
const MaxPartSize = 512 * 1024

// MaxGroupSize is the maximum number of media in one grouped message (album).
// Refer to https://core.telegram.org/method/messages.sendMultiMedia
const MaxGroupSize = 10

type Uploader struct {
	opts Options
}

type Options struct {
	Client   *tg.Client
	Threads  int
	Iter     Iter
	Progress Progress
}

func New(o Options) *Uploader {
	return &Uploader{opts: o}
}

// mediaOption is a media option that can be sent either as a single media
// message or as a member of an album.
type mediaOption interface {
	message.MediaOption
	message.MultiMediaOption
}

func (u *Uploader) Upload(ctx context.Context, limit int) error {
	wg, wgctx := errgroup.WithContext(ctx)
	wg.SetLimit(limit)

	for u.opts.Iter.Next(wgctx) {
		elem := u.opts.Iter.Value()

		wg.Go(func() (rerr error) {
			// grouped messages report progress per member file
			progressElems := []Elem{elem}
			if g, ok := elem.(GroupElem); ok {
				progressElems = g.Elems()
			}

			for _, e := range progressElems {
				u.opts.Progress.OnAdd(e)
			}
			defer func() {
				for _, e := range progressElems {
					u.opts.Progress.OnDone(e, rerr)
				}
			}()

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

func (u *Uploader) upload(ctx context.Context, elem Elem) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// a group with only one member is just sent as a single media message
	if g, ok := elem.(GroupElem); ok && len(g.Elems()) > 1 {
		return u.uploadGroup(ctx, g)
	}

	media, err := u.buildMedia(ctx, elem)
	if err != nil {
		return err
	}

	if _, err = u.sender(elem).Media(ctx, media); err != nil {
		return errors.Wrap(err, "send message")
	}

	return nil
}

// uploadGroup uploads all files of the group, then sends them as one grouped
// message (album). All members must share the same destination.
func (u *Uploader) uploadGroup(ctx context.Context, group GroupElem) error {
	elems := group.Elems()

	media := make([]message.MultiMediaOption, 0, len(elems))
	for _, elem := range elems {
		m, err := u.buildMedia(ctx, elem)
		if err != nil {
			return err
		}

		media = append(media, m)
	}

	_, err := u.sender(group).Album(ctx, media[0], media[1:]...)
	if err != nil {
		return errors.Wrap(err, "send album")
	}

	return nil
}

// buildMedia uploads the file bytes and returns a media option with proper
// attributes (photo/video/audio/document) and caption applied.
func (u *Uploader) buildMedia(ctx context.Context, elem Elem) (mediaOption, error) {
	up := uploader.NewUploader(u.opts.Client).
		WithPartSize(MaxPartSize).
		WithThreads(u.opts.Threads).
		WithProgress(&wrapProcess{
			elem:    elem,
			process: u.opts.Progress,
		})

	f, err := up.Upload(ctx, uploader.NewUpload(elem.File().Name(), elem.File(), elem.File().Size()))
	if err != nil {
		return nil, errors.Wrap(err, "upload file")
	}

	if _, err = elem.File().Seek(0, io.SeekStart); err != nil {
		return nil, errors.Wrap(err, "seek file")
	}
	mime, err := mimetype.DetectReader(elem.File())
	if err != nil {
		return nil, errors.Wrap(err, "detect mime")
	}

	// here convert underlying entities to formatters for message caption
	caption := styling.Custom(func(eb *entity.Builder) error {
		msg, entities := elem.Caption()
		eb.Format(msg, lo.Map(entities, func(item tg.MessageEntityClass, _ int) entity.Formatter {
			return func(_, _ int) tg.MessageEntityClass {
				return item
			}
		})...)
		return nil
	})

	doc := message.UploadedDocument(f, caption).MIME(mime.String()).Filename(elem.File().Name())
	// upload thumbnail TODO(iyear): maybe still unavailable
	if thumb, ok := elem.Thumb(); ok {
		if thumbFile, err := uploader.NewUploader(u.opts.Client).
			FromReader(ctx, thumb.Name(), thumb); err == nil {
			doc = doc.Thumb(thumbFile)
		}
	}

	var media mediaOption = doc

	switch {
	case mediautil.IsImage(mime.String()) && elem.AsPhoto():
		// webp should be uploaded as document
		if mime.String() == "image/webp" {
			break
		}
		// upload as photo
		media = message.UploadedPhoto(f, caption)
	case mediautil.IsVideo(mime.String()):
		// reset reader
		if _, err = elem.File().Seek(0, io.SeekStart); err != nil {
			return nil, errors.Wrap(err, "seek file")
		}
		if dur, w, h, err := mediautil.GetMP4Info(elem.File()); err == nil {
			// #132. There may be some errors, but we can still upload the file
			media = doc.Video().
				Duration(time.Duration(dur)*time.Second).
				Resolution(w, h).
				SupportsStreaming()
		}
	case mediautil.IsAudio(mime.String()):
		media = doc.Audio().Title(fsutil.GetNameWithoutExt(elem.File().Name()))
	}

	return media, nil
}

func (u *Uploader) sender(elem Elem) *message.Builder {
	up := uploader.NewUploader(u.opts.Client).
		WithPartSize(MaxPartSize).
		WithThreads(u.opts.Threads)

	return message.NewSender(u.opts.Client).
		WithUploader(up).
		To(elem.To()).
		Reply(elem.Thread())
}
