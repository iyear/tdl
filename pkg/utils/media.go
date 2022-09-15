package utils

import (
	"strings"
)

type media struct{}

var Media = media{}

func (m media) split(mime string) (primary string, sub string, ok bool) {
	types := strings.Split(mime, "/")

	if len(types) != 2 {
		return
	}

	return types[0], types[1], true
}

func (m media) IsVideo(mime string) bool {
	primary, _, ok := m.split(mime)

	return primary == "video" && ok
}

func (m media) IsAudio(mime string) bool {
	primary, _, ok := m.split(mime)

	return primary == "audio" && ok
}

func (m media) IsImage(mime string) bool {
	primary, _, ok := m.split(mime)

	return primary == "image" && ok
}
