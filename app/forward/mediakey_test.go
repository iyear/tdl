package forward

import (
	"testing"

	"github.com/gotd/td/tg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMediaKeyDocument(t *testing.T) {
	msg := &tg.Message{
		ID: 1,
		Media: &tg.MessageMediaDocument{
			Document: &tg.Document{
				ID:   42,
				Size: 1024,
			},
		},
	}

	key, ok := mediaKey(msg)
	require.True(t, ok)
	assert.Equal(t, "doc:42:1024", key)

	// same document id and size must produce the same key
	same := &tg.Message{ID: 2, Media: msg.Media}
	key2, ok := mediaKey(same)
	require.True(t, ok)
	assert.Equal(t, key, key2)

	// different size must produce a different key
	diffSize := &tg.Message{
		ID: 3,
		Media: &tg.MessageMediaDocument{
			Document: &tg.Document{ID: 42, Size: 2048},
		},
	}
	key3, ok := mediaKey(diffSize)
	require.True(t, ok)
	assert.NotEqual(t, key, key3)

	// different id must produce a different key
	diffID := &tg.Message{
		ID: 4,
		Media: &tg.MessageMediaDocument{
			Document: &tg.Document{ID: 43, Size: 1024},
		},
	}
	key4, ok := mediaKey(diffID)
	require.True(t, ok)
	assert.NotEqual(t, key, key4)
}

func TestMediaKeyPhoto(t *testing.T) {
	photo := &tg.Photo{ID: 100}

	msg := &tg.Message{
		ID:    5,
		Media: &tg.MessageMediaPhoto{Photo: photo},
	}

	key, ok := mediaKey(msg)
	require.True(t, ok)
	assert.Equal(t, "photo:100", key)

	// same photo must produce the same key regardless of message id
	same := &tg.Message{ID: 6, Media: &tg.MessageMediaPhoto{Photo: photo}}
	key2, ok := mediaKey(same)
	require.True(t, ok)
	assert.Equal(t, key, key2)
}

func TestMediaKeyNoMedia(t *testing.T) {
	cases := map[string]*tg.Message{
		"no media":      {ID: 7},
		"empty text":    {ID: 8, Message: ""},
		"unknown media": {ID: 9, Media: &tg.MessageMediaGeo{Geo: &tg.GeoPointEmpty{}}},
		"doc empty":     {ID: 10, Media: &tg.MessageMediaDocument{Document: &tg.DocumentEmpty{}}},
		"photo empty":   {ID: 11, Media: &tg.MessageMediaPhoto{Photo: &tg.PhotoEmpty{}}},
	}

	for name, msg := range cases {
		t.Run(name, func(t *testing.T) {
			key, ok := mediaKey(msg)
			assert.False(t, ok)
			assert.Empty(t, key)
		})
	}
}

func TestMediaDestKey(t *testing.T) {
	key := mediaDestKey(7, 0, "doc:42:1024")

	// same destination and media -> same key
	assert.Equal(t, key, mediaDestKey(7, 0, "doc:42:1024"))

	// different thread -> different key
	assert.NotEqual(t, key, mediaDestKey(7, 5, "doc:42:1024"))

	// different destination -> different key
	assert.NotEqual(t, key, mediaDestKey(8, 0, "doc:42:1024"))
}
