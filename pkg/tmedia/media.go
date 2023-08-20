package tmedia

import (
	"github.com/gotd/td/tg"

	"github.com/iyear/tdl/pkg/downloader"
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
