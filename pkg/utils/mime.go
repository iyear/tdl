package utils

import "strings"

type mime struct{}

var MIME = mime{}

func (m mime) split(mime string) (primary string, sub string, ok bool) {
	types := strings.Split(mime, "/")

	if len(types) != 2 {
		return
	}

	return types[0], types[1], true
}

func (m mime) IsVideo(mime string) bool {
	primary, _, ok := m.split(mime)

	return primary == "video" && ok
}

func (m mime) IsAudio(mime string) bool {
	primary, _, ok := m.split(mime)

	return primary == "audio" && ok
}

func (m mime) IsImage(mime string) bool {
	primary, _, ok := m.split(mime)

	return primary == "image" && ok
}
