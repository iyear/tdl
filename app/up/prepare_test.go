package up

import (
	"testing"

	"github.com/iyear/tdl/core/uploader"
)

func TestAlbumGroupID(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{path: "alice_2026-07-21_10-00-00.mp4", want: "alice_2026-07-21"},
		{path: "bob_2026-07-21_11-30-00.mov", want: "bob_2026-07-21"},
		{path: "random.mp4", want: ""},
	}

	for _, tt := range tests {
		if got := albumGroupID(tt.path); got != tt.want {
			t.Fatalf("albumGroupID(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestIsDetectVideoExt(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{path: "a.mp4", want: true},
		{path: "a.M4V", want: true},
		{path: "a.mov", want: true},
		{path: "a.mkv", want: false},
	}

	for _, tt := range tests {
		if got := isDetectVideoExt(tt.path); got != tt.want {
			t.Fatalf("isDetectVideoExt(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestDecideMediaType(t *testing.T) {
	opts := Options{Photo: true, Video: false, DetectVideo: true}

	if got := decideMediaType("x.mp4", "application/octet-stream", opts); got != uploader.MediaTypeVideo {
		t.Fatalf("detect-video expected video, got %s", got)
	}
	if got := decideMediaType("x.jpg", "image/jpeg", opts); got != uploader.MediaTypePhoto {
		t.Fatalf("photo expected photo, got %s", got)
	}
	if got := decideMediaType("x.bin", "application/octet-stream", opts); got != uploader.MediaTypeDocument {
		t.Fatalf("default expected document, got %s", got)
	}
}
