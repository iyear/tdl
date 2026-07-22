package uploader

import (
	"context"
	"math/rand"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/telegram/message/styling"
	tduploader "github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"
)

// MaxPartSize refer to https://core.telegram.org/api/files#uploading-files
const MaxPartSize = 512 * 1024

type Uploader struct {
	opts Options
	rand *rand.Rand
}

type Options struct {
	Client   *tg.Client
	Threads  int
	Iter     Iter
	Progress Progress
}

func New(o Options) *Uploader {
	return &Uploader{
		opts: o,
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (u *Uploader) Upload(ctx context.Context, limit int) error {
	wg, wgctx := errgroup.WithContext(ctx)
	wg.SetLimit(limit)

	for u.opts.Iter.Next(wgctx) {
		elem := u.opts.Iter.Value()

		wg.Go(func() error {
			u.opts.Progress.OnAdd(elem)

			err := u.upload(wgctx, elem)
			u.opts.Progress.OnDone(elem, err)

			if err != nil && errors.Is(err, context.Canceled) {
				return errors.Wrap(err, "upload")
			}
			return nil
		})
	}

	if err := u.opts.Iter.Err(); err != nil {
		return errors.Wrap(err, "iter")
	}

	return wg.Wait()
}

func (u *Uploader) upload(ctx context.Context, elem Elem) (rerr error) {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	media := elem.Media()
	if len(media) == 0 {
		return errors.New("empty media")
	}

	if len(media) == 1 {
		if err := u.sendSingle(ctx, elem, media[0]); err != nil {
			return err
		}
	} else {
		if err := u.sendAlbum(ctx, elem, media); err != nil {
			return err
		}
	}

	if elem.Remove() {
		for _, p := range elem.Paths() {
			if err := os.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) {
				return errors.Wrapf(err, "remove file: %s", p)
			}
		}
	}

	return nil
}

func (u *Uploader) sendSingle(ctx context.Context, elem Elem, media MediaDescriptor) error {
	ap := newAlbumProgress(elem, u.opts.Progress)
	inputMedia, _, err := u.buildInputMedia(ctx, elem, media, ap)
	if err != nil {
		return errors.Wrap(err, "build media")
	}

	caption := styling.Custom(func(eb *entity.Builder) error {
		eb.Format(media.Caption, lo.Map(elem.CaptionEntities(media), func(item tg.MessageEntityClass, _ int) entity.Formatter {
			return func(_, _ int) tg.MessageEntityClass {
				return item
			}
		})...)
		return nil
	})

	switch m := inputMedia.(type) {
	case *tg.InputMediaUploadedPhoto:
		_, err = message.NewSender(u.opts.Client).
			To(elem.To()).
			Reply(elem.Thread()).
			Media(ctx, message.UploadedPhoto(m.File, caption))
		if err != nil {
			return errors.Wrap(err, "send photo")
		}
	case *tg.InputMediaUploadedDocument:
		doc := message.UploadedDocument(m.File, caption).MIME(m.MimeType).Filename(filepath.Base(media.Path))
		if m.Thumb != nil {
			doc = doc.Thumb(m.Thumb)
		}
		var mediaOption message.MediaOption = doc
		if media.Type == MediaTypeVideo {
			mediaOption = doc.Video().
				SupportsStreaming().
				Duration(time.Duration(media.Duration)*time.Second).
				Resolution(media.Width, media.Height)
		}
		_, err = message.NewSender(u.opts.Client).
			To(elem.To()).
			Reply(elem.Thread()).
			Media(ctx, mediaOption)
		if err != nil {
			return errors.Wrap(err, "send document")
		}
	default:
		return errors.Errorf("unsupported media %T", inputMedia)
	}

	return nil
}

func (u *Uploader) sendAlbum(ctx context.Context, elem Elem, mediaList []MediaDescriptor) error {
	items := make([]tg.InputSingleMedia, len(mediaList))
	randomIDs := make([]int64, len(mediaList))
	for i := range randomIDs {
		randomIDs[i] = u.rand.Int63()
	}

	ap := newAlbumProgress(elem, u.opts.Progress)
	peer := elem.To()
	wg, wgctx := errgroup.WithContext(ctx)
	for idx, media := range mediaList {
		idx, media := idx, media
		wg.Go(func() error {
			uploaded, entities, err := u.buildInputMedia(wgctx, elem, media, ap)
			if err != nil {
				return errors.Wrap(err, "build album media")
			}

			committed, err := u.commitAlbumMedia(wgctx, peer, uploaded)
			if err != nil {
				return errors.Wrap(err, "commit album media")
			}

			item := tg.InputSingleMedia{
				Media:    committed,
				RandomID: randomIDs[idx],
				Message:  media.Caption,
				Entities: entities,
			}
			item.SetFlags()
			items[idx] = item
			return nil
		})
	}
	if err := wg.Wait(); err != nil {
		return err
	}

	req := &tg.MessagesSendMultiMediaRequest{
		Peer:       peer,
		ReplyTo:    getReplyTo(elem.Thread()),
		MultiMedia: items,
	}
	req.SetFlags()

	if _, err := u.opts.Client.MessagesSendMultiMedia(ctx, req); err != nil {
		return errors.Wrap(err, "send multi media")
	}

	return nil
}

func (u *Uploader) commitAlbumMedia(ctx context.Context, peer tg.InputPeerClass, media tg.InputMediaClass) (tg.InputMediaClass, error) {
	res, err := u.opts.Client.MessagesUploadMedia(ctx, &tg.MessagesUploadMediaRequest{
		Peer:  peer,
		Media: media,
	})
	if err != nil {
		return nil, errors.Wrap(err, "upload media")
	}
	return convertToInputMedia(res)
}

func convertToInputMedia(m tg.MessageMediaClass) (tg.InputMediaClass, error) {
	switch v := m.(type) {
	case *tg.MessageMediaPhoto:
		photo, ok := v.Photo.AsNotEmpty()
		if !ok {
			return nil, errors.Errorf("unexpected photo type %T", v.Photo)
		}
		return &tg.InputMediaPhoto{
			ID:         photo.AsInput(),
			TTLSeconds: v.TTLSeconds,
		}, nil
	case *tg.MessageMediaDocument:
		document, ok := v.Document.AsNotEmpty()
		if !ok {
			return nil, errors.Errorf("unexpected document type %T", v.Document)
		}
		return &tg.InputMediaDocument{
			ID:         document.AsInput(),
			TTLSeconds: v.TTLSeconds,
		}, nil
	default:
		return nil, errors.Errorf("unexpected media %T", m)
	}
}

func (u *Uploader) buildInputMedia(ctx context.Context, elem Elem, media MediaDescriptor, ap *albumProgress) (tg.InputMediaClass, []tg.MessageEntityClass, error) {
	entities := elem.CaptionEntities(media)

	file, size, err := openSized(media.Path)
	if err != nil {
		return nil, nil, errors.Wrap(err, "open media file")
	}
	defer func() { _ = file.Close() }()

	main := tduploader.NewUploader(u.opts.Client).
		WithPartSize(MaxPartSize).
		WithThreads(u.opts.Threads).
		WithProgress(ap.newFileProgress())

	inputFile, err := main.Upload(ctx, tduploader.NewUpload(filepath.Base(media.Path), file, size))
	if err != nil {
		return nil, nil, errors.Wrap(err, "upload media")
	}

	switch media.Type {
	case MediaTypePhoto:
		m := &tg.InputMediaUploadedPhoto{File: inputFile}
		m.SetFlags()
		return m, entities, nil
	case MediaTypeVideo, MediaTypeDocument:
		doc := &tg.InputMediaUploadedDocument{
			NosoundVideo: false,
			File:         inputFile,
			MimeType:     media.Mime,
			Attributes:   buildDocumentAttributes(media),
		}

		if media.Thumb != "" {
			thumb, thumbSize, err := openSized(media.Thumb)
			if err == nil {
				defer func() { _ = thumb.Close() }()
				if thumbFile, upErr := tduploader.NewUploader(u.opts.Client).
					WithPartSize(MaxPartSize).
					WithThreads(u.opts.Threads).
					Upload(ctx, tduploader.NewUpload(filepath.Base(media.Thumb), thumb, thumbSize)); upErr == nil {
					doc.Thumb = thumbFile
				}
			}
		}

		doc.SetFlags()
		return doc, entities, nil
	default:
		return nil, nil, errors.Errorf("unsupported media type: %s", media.Type)
	}
}

func buildDocumentAttributes(media MediaDescriptor) []tg.DocumentAttributeClass {
	attrs := []tg.DocumentAttributeClass{
		&tg.DocumentAttributeFilename{
			FileName: filepath.Base(media.Path),
		},
	}

	if media.Type == MediaTypeVideo {
		v := &tg.DocumentAttributeVideo{
			RoundMessage:      false,
			SupportsStreaming: media.Streaming,
			Nosound:           false,
			W:                 media.Width,
			H:                 media.Height,
			Duration:          float64(media.Duration),
		}
		v.SetFlags()
		attrs = append(attrs, v)
	}

	return attrs
}

func openSized(path string) (*os.File, int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}

	stat, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, 0, err
	}

	return f, stat.Size(), nil
}

type albumProgress struct {
	elem     Elem
	total    int64
	uploaded atomic.Int64
	process  Progress
}

func newAlbumProgress(elem Elem, process Progress) *albumProgress {
	return &albumProgress{
		elem:    elem,
		total:   elem.TotalSize(),
		process: process,
	}
}

func (a *albumProgress) newFileProgress() *fileProgress {
	return &fileProgress{album: a}
}

type fileProgress struct {
	album *albumProgress
	last  int64
}

func (p *fileProgress) Chunk(_ context.Context, state tduploader.ProgressState) error {
	delta := state.Uploaded - p.last
	p.last = state.Uploaded

	total := p.album.uploaded.Add(delta)
	p.album.process.OnUpload(p.album.elem, ProgressState{
		Uploaded: total,
		Total:    p.album.total,
	})
	return nil
}

func getReplyTo(msgID int) tg.InputReplyToClass {
	if msgID == 0 {
		return nil
	}
	reply := &tg.InputReplyToMessage{ReplyToMsgID: msgID}
	reply.SetFlags()
	return reply
}
