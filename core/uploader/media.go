package uploader

import (
	"io"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/go-faster/errors"

	"github.com/iyear/tdl/core/util/mediautil"
)

type mediaKind int

const (
	mediaDocument mediaKind = iota
	mediaPhoto
	mediaVideo
	mediaAudio
)

// MediaInfo describes how a file will be sent to Telegram.
type MediaInfo struct {
	mime          string
	kind          mediaKind
	duration      time.Duration
	width, height int
}

// MIME returns the MIME type detected from the file contents.
func (m MediaInfo) MIME() string {
	return m.mime
}

// CanGroupWith reports whether both files can be sent in the same album.
func (m MediaInfo) CanGroupWith(other MediaInfo) bool {
	return m.kind == other.kind ||
		(m.kind == mediaPhoto && other.kind == mediaVideo) ||
		(m.kind == mediaVideo && other.kind == mediaPhoto)
}

// InspectMedia determines the send type and rewinds the file for upload.
func InspectMedia(file io.ReadSeeker, asPhoto bool) (MediaInfo, error) {
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return MediaInfo{}, errors.Wrap(err, "seek file")
	}
	mime, err := mimetype.DetectReader(file)
	if err != nil {
		return MediaInfo{}, errors.Wrap(err, "detect mime")
	}

	info := MediaInfo{mime: mime.String(), kind: mediaDocument}
	switch {
	case asPhoto && mediautil.IsImage(info.mime) && info.mime != "image/webp":
		info.kind = mediaPhoto
	case mediautil.IsVideo(info.mime):
		if _, err = file.Seek(0, io.SeekStart); err != nil {
			return MediaInfo{}, errors.Wrap(err, "seek video")
		}
		if duration, width, height, err := mediautil.GetMP4Info(file); err == nil {
			info.kind = mediaVideo
			info.duration = time.Duration(duration) * time.Second
			info.width, info.height = width, height
		}
	case mediautil.IsAudio(info.mime):
		info.kind = mediaAudio
	}

	if _, err = file.Seek(0, io.SeekStart); err != nil {
		return MediaInfo{}, errors.Wrap(err, "rewind file")
	}
	return info, nil
}
