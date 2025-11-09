package tmedia

import (
	"fmt"
	"strconv"

	"github.com/gabriel-vasile/mimetype"
	"github.com/go-faster/errors"
	"github.com/gotd/td/tg"
)

func GetDocumentInfo(doc *tg.MessageMediaDocument) (*Media, error) {
	d, ok := doc.Document.(*tg.Document)
	if !ok {
		return nil, errors.New(fmt.Sprintf("expected *tg.Document, got %T", doc.Document))
	}

	return &Media{
		InputFileLoc: &tg.InputDocumentFileLocation{
			ID:            d.ID,
			AccessHash:    d.AccessHash,
			FileReference: d.FileReference,
		},
		Name: GetDocumentName(d),
		Size: d.Size,
		DC:   d.DCID,
		Date: int64(d.Date),
	}, nil
}

func GetDocumentName(doc *tg.Document) string {
	for _, attr := range doc.Attributes {
		name, ok := attr.(*tg.DocumentAttributeFilename)
		if ok {
			return name.FileName
		}
	}

	// #185: stable file name so --skip-same can work
	mime := mimetype.Lookup(doc.MimeType)
	ext := ".unknown"
	if mime != nil {
		ext = mime.Extension()
	}
	return strconv.FormatInt(doc.ID, 10) + ext
}
