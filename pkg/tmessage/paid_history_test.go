package tmessage

import (
	"testing"

	"github.com/gotd/td/tg"
	"github.com/stretchr/testify/require"
)

func TestUnlockedPaidMediaCount(t *testing.T) {
	withMedia := func(media tg.MessageMediaClass) *tg.Message {
		message := &tg.Message{Media: media}
		message.SetFlags()
		return message
	}
	photo := func(id int64) *tg.MessageExtendedMedia {
		return &tg.MessageExtendedMedia{Media: &tg.MessageMediaPhoto{Photo: &tg.Photo{
			ID: id,
			Sizes: []tg.PhotoSizeClass{&tg.PhotoSize{
				Type: "x",
				Size: 1,
			}},
		}}}
	}

	tests := []struct {
		name    string
		message *tg.Message
		want    int
	}{
		{name: "no media", message: &tg.Message{}, want: 0},
		{name: "ordinary media", message: withMedia(&tg.MessageMediaPhoto{}), want: 0},
		{
			name: "unlocked and preview items",
			message: withMedia(&tg.MessageMediaPaidMedia{
				ExtendedMedia: []tg.MessageExtendedMediaClass{
					photo(1),
					&tg.MessageExtendedMediaPreview{},
					photo(2),
				},
			}),
			want: 2,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.want, unlockedPaidMediaCount(test.message))
		})
	}
}

func TestDialogMediaCount(t *testing.T) {
	dialog := &Dialog{MediaCounts: map[int]int{10: 4}}
	require.Equal(t, 4, dialog.MediaCount(10))
	require.Equal(t, 1, dialog.MediaCount(11))
}
