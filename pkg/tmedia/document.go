package tmedia

import (
	"strconv"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gotd/td/tg"
)

func GetDocumentInfo(doc *tg.MessageMediaDocument) (*Media, bool) {
	d, ok := doc.Document.(*tg.Document)
	if !ok {
		return nil, false
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
	}, true
}

func GetDocumentName(doc *tg.Document) string {
	for _, attr := range doc.Attributes {
		name, ok := attr.(*tg.DocumentAttributeFilename)
		if ok {
			return name.FileName
		}
	}

	// #185: stable file name so --skip-same can work
	ext := mimetype.Lookup(doc.MimeType).Extension()
	return strconv.FormatInt(doc.ID, 10) + ext
}
