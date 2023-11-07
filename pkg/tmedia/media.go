package tmedia

import (
	"github.com/gotd/td/tg"
)

type Media struct {
	InputFileLoc tg.InputFileLocationClass
	Name         string
	Size         int64
	DC           int
}

func GetMedia(msg tg.MessageClass) (*Media, bool) {
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
