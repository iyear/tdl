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
	"github.com/iyear/tdl/core/util/thumbnail"
)

// MaxPartSize refer to https://core.telegram.org/api/files#uploading-files
const MaxPartSize = 512 * 1024

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

func (u *Uploader) Upload(ctx context.Context, limit int) error {
	wg, wgctx := errgroup.WithContext(ctx)
	wg.SetLimit(limit)

	for u.opts.Iter.Next(wgctx) {
		elem := u.opts.Iter.Value()

		wg.Go(func() (rerr error) {
			u.opts.Progress.OnAdd(elem)
			defer func() { u.opts.Progress.OnDone(elem, rerr) }()

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

	up := uploader.NewUploader(u.opts.Client).
		WithPartSize(MaxPartSize).
		WithThreads(u.opts.Threads).
		WithProgress(&wrapProcess{
			elem:    elem,
			process: u.opts.Progress,
		})

	f, err := up.Upload(ctx, uploader.NewUpload(elem.File().Name(), elem.File(), elem.File().Size()))
	if err != nil {
		return errors.Wrap(err, "upload file")
	}

	if _, err = elem.File().Seek(0, io.SeekStart); err != nil {
		return errors.Wrap(err, "seek file")
	}
	mime, err := mimetype.DetectReader(elem.File())
	if err != nil {
		return errors.Wrap(err, "detect mime")
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
	isVideoUpload := mediautil.IsVideo(mime.String()) && elem.AsVideo()
	if thumb, ok := elem.Thumb(); ok && !isVideoUpload {
		if thumbFile, err := uploader.NewUploader(u.opts.Client).
			FromReader(ctx, thumb.Name(), thumb); err == nil {
			doc = doc.Thumb(thumbFile)
		}
	}

	var media message.MediaOption = doc

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
			return errors.Wrap(err, "seek file")
		}
		dur, w, h, videoErr := mediautil.GetMP4Info(elem.File())
		if videoErr != nil && elem.AsVideo() {
			if _, err = elem.File().Seek(0, io.SeekStart); err != nil {
				return errors.Wrap(err, "seek file")
			}
			dur, w, h, videoErr = mediautil.GetMatroskaInfo(elem.File())
		}

		if videoErr == nil && elem.AsVideo() {
			attributes := []tg.DocumentAttributeClass{
				&tg.DocumentAttributeFilename{FileName: elem.File().Name()},
				&tg.DocumentAttributeVideo{
					Duration:          float64(dur),
					W:                 w,
					H:                 h,
					SupportsStreaming: true,
				},
			}
			document := &tg.InputMediaUploadedDocument{
				File:       f,
				MimeType:   mime.String(),
				Attributes: attributes,
			}

			thumb, hasThumb := elem.Thumb()
			if !hasThumb {
				thumb, err = thumbnail.NewBlack(w, h)
				if err != nil {
					return errors.Wrap(err, "create default video cover")
				}
			}
			thumbFile, err := u.uploadThumbnail(ctx, thumb)
			if err != nil {
				return errors.Wrap(err, "upload video thumbnail")
			}
			document.Thumb = thumbFile

			if _, err = thumb.Seek(0, io.SeekStart); err != nil {
				return errors.Wrap(err, "seek video cover")
			}
			cover, err := u.uploadVideoCover(ctx, thumb)
			if err != nil {
				return errors.Wrap(err, "upload video cover")
			}
			document.VideoCover = cover
			document.SetFlags()
			media = message.Media(document, caption)
		} else if videoErr == nil {
			// #132. There may be some errors, but we can still upload the file
			media = doc.Video().
				Duration(time.Duration(dur)*time.Second).
				Resolution(w, h).
				SupportsStreaming()
		} else if elem.AsVideo() {
			return errors.Wrap(videoErr, "probe MP4 or MKV video metadata")
		}
	case mediautil.IsAudio(mime.String()):
		media = doc.Audio().Title(fsutil.GetNameWithoutExt(elem.File().Name()))
	}

	_, err = message.NewSender(u.opts.Client).
		WithUploader(up).
		To(elem.To()).
		Reply(elem.Thread()).
		Media(ctx, media)
	if err != nil {
		return errors.Wrap(err, "send message")
	}

	return nil
}

func (u *Uploader) uploadThumbnail(ctx context.Context, thumb File) (tg.InputFileClass, error) {
	return uploader.NewUploader(u.opts.Client).
		WithPartSize(MaxPartSize).
		WithThreads(u.opts.Threads).
		Upload(ctx, uploader.NewUpload(thumb.Name(), thumb, thumb.Size()))
}

func (u *Uploader) uploadVideoCover(ctx context.Context, thumb File) (tg.InputPhotoClass, error) {
	up := uploader.NewUploader(u.opts.Client).
		WithPartSize(MaxPartSize).
		WithThreads(u.opts.Threads)
	thumbFile, err := up.Upload(ctx, uploader.NewUpload(thumb.Name(), thumb, thumb.Size()))
	if err != nil {
		return nil, errors.Wrap(err, "upload cover file")
	}
	response, err := u.opts.Client.MessagesUploadMedia(ctx, &tg.MessagesUploadMediaRequest{
		Peer:  &tg.InputPeerSelf{},
		Media: &tg.InputMediaUploadedPhoto{File: thumbFile},
	})
	if err != nil {
		return nil, errors.Wrap(err, "convert cover to photo")
	}
	mediaPhoto, ok := response.(*tg.MessageMediaPhoto)
	if !ok {
		return nil, errors.Errorf("unexpected video cover response: %T", response)
	}
	photo, ok := mediaPhoto.GetPhoto()
	if !ok {
		return nil, errors.New("video cover photo is empty")
	}
	photoObject, ok := photo.(*tg.Photo)
	if !ok {
		return nil, errors.Errorf("unexpected video cover photo: %T", photo)
	}
	return &tg.InputPhoto{
		ID:            photoObject.ID,
		AccessHash:    photoObject.AccessHash,
		FileReference: photoObject.FileReference,
	}, nil
}
