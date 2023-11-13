package tmedia

import (
	"github.com/gotd/td/tg"
)

func ConvInputMedia(media tg.MessageMediaClass) (tg.InputMediaClass, bool) {
	switch v := media.(type) {
	case *tg.MessageMediaPhoto:
		return ConvInputMediaPhoto(v)
	case *tg.MessageMediaGeo:
		return ConvInputMediaGeo(v)
	case *tg.MessageMediaContact:
		return ConvInputMediaContact(v)
	case *tg.MessageMediaDocument:
		return ConvInputMediaDocument(v)
	case *tg.MessageMediaVenue:
		return ConvInputMediaVenue(v)
	case *tg.MessageMediaGame:
		return ConvInputMediaGame(v)
	case *tg.MessageMediaInvoice:
		return ConvInputMediaInvoice(v)
	case *tg.MessageMediaGeoLive:
		return ConvInputMediaGeoLive(v)
	case *tg.MessageMediaPoll:
		return ConvInputMediaPoll(v)
	case *tg.MessageMediaDice:
		return ConvInputMediaDice(v)
	case *tg.MessageMediaStory:
		return ConvInputMediaStory(v)
	case *tg.MessageMediaUnsupported:
		return nil, false
	default:
		return nil, false
	}
}

func ConvInputMediaPhoto(v *tg.MessageMediaPhoto) (*tg.InputMediaPhoto, bool) {
	switch t := v.Photo.(type) {
	case *tg.PhotoEmpty:
		return nil, false
	case *tg.Photo:
		p := &tg.InputPhoto{}
		p.FillFrom(t)

		ret := &tg.InputMediaPhoto{
			Spoiler:    v.Spoiler,
			ID:         p,
			TTLSeconds: v.TTLSeconds,
		}
		ret.SetFlags()
		return ret, true
	default:
		return nil, false
	}
}

func ConvInputMediaGeo(v *tg.MessageMediaGeo) (*tg.InputMediaGeoPoint, bool) {
	switch t := v.Geo.(type) {
	case *tg.GeoPointEmpty:
		return nil, false
	case *tg.GeoPoint:
		g := &tg.InputGeoPoint{}
		g.FillFrom(t)
		g.SetFlags()

		return &tg.InputMediaGeoPoint{
			GeoPoint: g,
		}, true
	default:
		return nil, false
	}
}

func ConvInputMediaContact(v *tg.MessageMediaContact) (*tg.InputMediaContact, bool) {
	c := &tg.InputMediaContact{}
	c.FillFrom(v)

	return c, true
}

func ConvInputMediaDocument(v *tg.MessageMediaDocument) (*tg.InputMediaDocument, bool) {
	switch t := v.Document.(type) {
	case *tg.DocumentEmpty:
		return nil, false
	case *tg.Document:
		d := &tg.InputDocument{}
		d.FillFrom(t)
		ret := &tg.InputMediaDocument{
			Spoiler:    v.Spoiler,
			ID:         d,
			TTLSeconds: v.TTLSeconds,
		}
		ret.SetFlags()

		return ret, true
	default:
		return nil, false
	}
}

func ConvInputMediaVenue(v *tg.MessageMediaVenue) (*tg.InputMediaVenue, bool) {
	geo, ok := ConvInputMediaGeo(&tg.MessageMediaGeo{Geo: v.Geo})
	if !ok {
		return nil, false
	}

	return &tg.InputMediaVenue{
		GeoPoint:  geo.GeoPoint,
		Title:     v.Title,
		Address:   v.Address,
		Provider:  v.Provider,
		VenueID:   v.VenueID,
		VenueType: v.VenueType,
	}, true
}

func ConvInputMediaGame(v *tg.MessageMediaGame) (*tg.InputMediaGame, bool) {
	g := &tg.InputGameID{}
	g.FillFrom(&v.Game)

	return &tg.InputMediaGame{
		ID: g,
	}, true
}

func ConvInputMediaInvoice(v *tg.MessageMediaInvoice) (*tg.InputMediaInvoice, bool) {
	// TODO(iyear): unsupported
	_ = v
	return nil, false
}

func ConvInputMediaGeoLive(v *tg.MessageMediaGeoLive) (*tg.InputMediaGeoLive, bool) {
	// TODO(): unsupported
	_ = v
	return nil, false
}

func ConvInputMediaPoll(v *tg.MessageMediaPoll) (*tg.InputMediaPoll, bool) {
	// TODO(): unsupported
	_ = v
	return nil, false
}

func ConvInputMediaDice(v *tg.MessageMediaDice) (*tg.InputMediaDice, bool) {
	return &tg.InputMediaDice{
		Emoticon: v.Emoticon,
	}, true
}

func ConvInputMediaStory(v *tg.MessageMediaStory) (*tg.InputMediaStory, bool) {
	// TODO(): unsupported
	_ = v
	return nil, false
}
