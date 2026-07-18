package tmedia

import (
	"testing"

	"github.com/gotd/td/tg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPaidMedia(t *testing.T) {
	paid := &tg.MessageMediaPaidMedia{
		StarsAmount: 10,
		ExtendedMedia: []tg.MessageExtendedMediaClass{
			paidPhoto(101),
			&tg.MessageExtendedMediaPreview{W: 320, H: 240},
			paidPhoto(202),
		},
	}

	media := GetPaidMedia(paid)
	require.Len(t, media, 3)
	require.NotNil(t, media[0])
	assert.Equal(t, "101.jpg", media[0].Name)
	assert.Nil(t, media[1], "locked previews must preserve their position")
	require.NotNil(t, media[2])
	assert.Equal(t, "202.jpg", media[2].Name)
}

func TestGetPaidMediaNil(t *testing.T) {
	assert.Nil(t, GetPaidMedia(nil))
}

func paidPhoto(id int64) *tg.MessageExtendedMedia {
	return &tg.MessageExtendedMedia{
		Media: &tg.MessageMediaPhoto{
			Photo: &tg.Photo{
				ID:            id,
				AccessHash:    id + 1,
				FileReference: []byte{1, 2, 3},
				Date:          1,
				Sizes: []tg.PhotoSizeClass{
					&tg.PhotoSize{Type: "x", W: 800, H: 600, Size: 1024},
				},
				DCID: 2,
			},
		},
	}
}
