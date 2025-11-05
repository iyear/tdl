package tmedia

import (
	"context"
	"fmt"

	"github.com/go-faster/errors"
	"github.com/gotd/td/tg"
)

type Media struct {
	InputFileLoc tg.InputFileLocationClass // mtproto file location of the media file
	Name         string                    // file name
	Size         int64                     // size in bytes
	DC           int                       // which DC the media is stored
	Date         int64                     // media creation(upload) timestamp
}

func ExtractMedia(ctx context.Context, m tg.MessageMediaClass) (tmedia *Media, isSupportedMediaType bool, err error) {
	switch m := m.(type) {
	case *tg.MessageMediaPhoto:
		tmedia, err = GetPhotoInfo(m)
		return tmedia, true, err
	case *tg.MessageMediaDocument:
		tmedia, err = GetDocumentInfo(m)
		return tmedia, true, err
	case *tg.MessageMediaInvoice:
		extendedMedia, ok := m.GetExtendedMedia()
		if !ok {
			return nil, true, errors.New("Could not extract extended media from *tg.MessageMediaInvoice")
		}
		return GetExtendedMedia(ctx, extendedMedia)
	}

	return nil, false, nil
}

func GetMedia(ctx context.Context, msg tg.MessageClass) (tmedia *Media, isSupportedMediaType bool, err error) {
	mm, ok := msg.(*tg.Message)
	if !ok {
		return nil, false, errors.Errorf("expected *tg.Message, got %T", msg)
	}

	media, ok := mm.GetMedia()
	if !ok {
		return nil, false, nil
	}

	return ExtractMedia(ctx, media)
}

func GetExtendedMedia(ctx context.Context, mm tg.MessageExtendedMediaClass) (tmedia *Media, isSupportedMediaType bool, err error) {
	m, ok := mm.(*tg.MessageExtendedMedia)
	if !ok {
		return nil, true, errors.Errorf(fmt.Sprintf("expected *tg.MessageExtendedMedia, got %T", mm))
	}
	return ExtractMedia(ctx, m.Media)
}

func GetDocumentThumb(doc *tg.Document) (*Media, error) {
	thumbs, exists := doc.GetThumbs()
	if !exists {
		return nil, errors.New("Could not extract thumbs from *tg.Document")
	}

	photoSize := &tg.PhotoSize{}
	for _, t := range thumbs {
		if p, ok := t.(*tg.PhotoSize); ok {
			photoSize = p
			break
		}
	}

	if photoSize == nil {
		return nil, errors.New("Could not extract photo size from *tg.Document")
	}

	return &Media{
		InputFileLoc: &tg.InputDocumentFileLocation{
			ID:            doc.ID,
			AccessHash:    doc.AccessHash,
			FileReference: doc.FileReference,
			ThumbSize:     photoSize.Type,
		},
		Name: "thumb.jpg",
		Size: int64(photoSize.Size),
		DC:   doc.DCID,
		Date: int64(doc.Date),
	}, nil
}
