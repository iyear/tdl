package tmedia

import (
	"strconv"

	"github.com/gotd/td/tg"
)

func GetPhotoInfo(photo *tg.MessageMediaPhoto) (*Media, bool) {
	p, ok := photo.Photo.(*tg.Photo)
	if !ok {
		return nil, false
	}

	tp, size, ok := GetPhotoSize(p.Sizes)
	if !ok {
		return nil, false
	}
	return &Media{
		InputFileLoc: &tg.InputPhotoFileLocation{
			ID:            p.ID,
			AccessHash:    p.AccessHash,
			FileReference: p.FileReference,
			ThumbSize:     tp,
		},
		// Telegram photo is compressed, and extension is always jpg.
		Name: strconv.FormatInt(p.ID, 10) + ".jpg", // unique name
		Size: int64(size),
		DC:   p.DCID,
	}, true
}

func GetPhotoSize(sizes []tg.PhotoSizeClass) (string, int, bool) {
	size := sizes[len(sizes)-1]
	switch s := size.(type) {
	case *tg.PhotoSize:
		return s.Type, s.Size, true
	case *tg.PhotoSizeProgressive:
		return s.Type, s.Sizes[len(s.Sizes)-1], true
	}

	return "", 0, false
}
