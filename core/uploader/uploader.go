package uploader

import (
	"context"

	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/iyear/tdl/core/logctx"
	"github.com/iyear/tdl/core/util/fsutil"
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

		wg.Go(func() error {
			elems := []Elem{elem}
			g, grouped := elem.(GroupElem)
			if grouped {
				elems = g.Elems()
			}

			for _, e := range elems {
				u.opts.Progress.OnAdd(e)
			}

			sendErr := u.upload(wgctx, elems)
			canceled := errors.Is(sendErr, context.Canceled)
			// Each member of a grouped upload shares the send result.
			var progressErr error
			if grouped || canceled {
				progressErr = sendErr
			}
			for _, e := range elems {
				u.opts.Progress.OnDone(e, progressErr)
			}

			if canceled {
				return errors.Wrap(sendErr, "upload")
			}
			return nil
		})
	}

	workerErr := wg.Wait()
	if err := u.opts.Iter.Err(); err != nil {
		return errors.Wrap(err, "iter")
	}
	return workerErr
}

func (u *Uploader) upload(ctx context.Context, elems []Elem) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	switch len(elems) {
	case 0:
		return errors.New("empty upload group")
	case 1:
	default:
		return u.uploadGroup(ctx, elems)
	}

	elem := elems[0]
	media, err := u.buildMedia(ctx, elem, false)
	if err != nil {
		return err
	}

	if _, err = u.sender(elem).Media(ctx, media); err != nil {
		return errors.Wrap(err, "send message")
	}

	return nil
}

// uploadGroup uploads all the members, then sends them in one album.
func (u *Uploader) uploadGroup(ctx context.Context, elems []Elem) error {
	media := make([]message.MultiMediaOption, 0, len(elems))
	for _, elem := range elems {
		m, err := u.buildMedia(ctx, elem, true)
		if err != nil {
			return err
		}

		media = append(media, m)
	}

	_, err := u.sender(elems[0]).Album(ctx, media[0], media[1:]...)
	if err != nil {
		return errors.Wrap(err, "send album")
	}

	return nil
}

// buildMedia uploads the file bytes and returns a media option with proper
// attributes (photo/video/audio/document) and caption applied.
func (u *Uploader) buildMedia(ctx context.Context, elem Elem, inAlbum bool) (mediaOption, error) {
	file := elem.File()
	var info MediaInfo
	if prepared, ok := elem.(interface{ MediaInfo() MediaInfo }); ok {
		info = prepared.MediaInfo()
	} else {
		var err error
		info, err = InspectMedia(file, elem.AsPhoto())
		if err != nil {
			return nil, err
		}
	}

	up := uploader.NewUploader(u.opts.Client).
		WithPartSize(MaxPartSize).
		WithThreads(u.opts.Threads).
		WithProgress(&wrapProcess{
			elem:    elem,
			process: u.opts.Progress,
		})

	f, err := up.Upload(ctx, uploader.NewUpload(file.Name(), file, file.Size()))
	if err != nil {
		return nil, errors.Wrap(err, "upload file")
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

	if info.kind == mediaPhoto {
		return message.UploadedPhoto(f, caption), nil
	}

	doc := message.UploadedDocument(f, caption).MIME(info.mime).Filename(file.Name())
	// The known size lets gotd upload the thumbnail as a regular InputFile.
	if thumb, ok := elem.Thumb(); ok {
		thumbFile, err := uploader.NewUploader(u.opts.Client).
			Upload(ctx, uploader.NewUpload(thumb.Name(), thumb, thumb.Size()))
		if err != nil {
			logctx.From(ctx).Warn("upload thumbnail", zap.String("thumb", thumb.Name()), zap.Error(err))
		} else {
			doc = doc.Thumb(thumbFile)
		}
	}

	switch info.kind {
	case mediaVideo:
		// Keep silent videos from being treated as animations in an album.
		return doc.NosoundVideo(inAlbum).Video().
			Duration(info.duration).
			Resolution(info.width, info.height).
			SupportsStreaming(), nil
	case mediaAudio:
		return doc.Audio().Title(fsutil.GetNameWithoutExt(file.Name())), nil
	default:
		// Keep Telegram from detecting another media type in a document album.
		return doc.ForceFile(inAlbum), nil
	}
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
