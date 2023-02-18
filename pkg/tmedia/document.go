package tmedia

import (
	"github.com/gotd/td/tg"
	"github.com/iyear/tdl/pkg/downloader"
	"strconv"
	"time"
)

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
