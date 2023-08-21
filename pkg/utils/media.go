package utils

import (
	"fmt"
	"io"
	"strings"

	"github.com/yapingcat/gomedia/go-mp4"
)

type media struct{}

var Media = media{}

func (m media) split(mime string) (primary string, sub string, ok bool) {
	types := strings.Split(mime, "/")

	if len(types) != 2 {
		return "", "", false
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

// GetMP4Info returns duration, width, height, error
func (m media) GetMP4Info(r io.ReadSeeker) (int, int, int, error) {
	d := mp4.CreateMp4Demuxer(r)

	tracks, err := d.ReadHead()
	if err != nil {
		return 0, 0, 0, err
	}

	for _, track := range tracks {
		if track.Cid == mp4.MP4_CODEC_H264 {
			info := d.GetMp4Info()
			return int(info.Duration / info.Timescale), int(track.Width), int(track.Height), nil
		}
	}

	return 0, 0, 0, fmt.Errorf("no h264 track found")
}
