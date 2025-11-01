package tmedia

import (
	"fmt"
	"strconv"

	"github.com/go-faster/errors"
	"github.com/gotd/td/tg"
)

func GetPhotoInfo(photo *tg.MessageMediaPhoto) (*Media, error) {
	p, ok := photo.Photo.(*tg.Photo)
	if !ok {
		return nil, errors.New(fmt.Sprintf("expected *tg.Photo, got %T", photo.Photo))
	}

	tp, size, err := GetPhotoSize(p.Sizes)
	if err != nil {
		return nil, err
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
		Date: int64(p.Date),
	}, nil
}

func GetPhotoSize(sizes []tg.PhotoSizeClass) (string, int, error) {
	size := sizes[len(sizes)-1]
	switch s := size.(type) {
	case *tg.PhotoSize:
		return s.Type, s.Size, nil
	case *tg.PhotoSizeProgressive:
		return s.Type, s.Sizes[len(s.Sizes)-1], nil

	}

	return "", 0, errors.New(fmt.Sprintf("unsupported photo size type: %T", size))
}
