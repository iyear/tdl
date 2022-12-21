package dl

import (
	"github.com/gotd/td/tg"
	"github.com/iyear/tdl/pkg/downloader"
	"sort"
	"strconv"
	"time"
)

func GetMedia(msg tg.MessageClass) (*downloader.Item, bool) {
	mm, ok := msg.(*tg.Message)
	if !ok {
		return nil, false
	}

	media, ok := mm.GetMedia()
	if !ok {
		return nil, false
	}

	switch m := media.(type) {
	case *tg.MessageMediaPhoto: // messageMediaPhoto#695150d7
		return GetPhotoInfo(m)
	case *tg.MessageMediaDocument: // messageMediaDocument#9cb070d7
		return GetDocumentInfo(m)
	}
	return nil, false
}

func GetPhotoInfo(photo *tg.MessageMediaPhoto) (*downloader.Item, bool) {
	p, ok := photo.Photo.(*tg.Photo)
	if !ok {
		return nil, false
	}

	tp, size, ok := GetPhotoSize(p)
	if !ok {
		return nil, false
	}
	return &downloader.Item{
		InputFileLoc: &tg.InputPhotoFileLocation{
			ID:            p.ID,
			AccessHash:    p.AccessHash,
			FileReference: p.FileReference,
			ThumbSize:     tp,
		},
		// Telegram photo is compressed, and extension is always jpg.
		Name: "photo.jpg",
		Size: size,
		DC:   p.DCID,
	}, true
}

func GetPhotoSize(photo *tg.Photo) (string, int64, bool) {
	for _, size := range photo.Sizes {
		s, ok := size.(*tg.PhotoSizeProgressive)
		if ok {
			sort.Ints(s.Sizes)
			return s.Type, int64(s.Sizes[len(s.Sizes)-1]), true
		}
		// TODO: old photo message only have PhotoSize type
	}
	return "", 0, false
}

func GetDocumentInfo(doc *tg.MessageMediaDocument) (*downloader.Item, bool) {
	d, ok := doc.Document.(*tg.Document)
	if !ok {
		return nil, false
	}

	return &downloader.Item{
		InputFileLoc: &tg.InputDocumentFileLocation{
			ID:            d.ID,
			AccessHash:    d.AccessHash,
			FileReference: d.FileReference,
		},
		Name: GetDocumentName(d.Attributes),
		Size: d.Size,
		DC:   d.DCID,
	}, true
}

func GetDocumentName(attrs []tg.DocumentAttributeClass) string {
	for _, attr := range attrs {
		name, ok := attr.(*tg.DocumentAttributeFilename)
		if ok {
			return name.FileName
		}
	}

	return strconv.FormatInt(time.Now().Unix(), 10)
}
