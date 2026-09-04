package up

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWalkPairsThumbnailWithoutTrimmingFilenameCharacters(t *testing.T) {
	dir := t.TempDir()
	video := filepath.Join(dir, "clip4.mp4")
	thumb := filepath.Join(dir, "clip4.thumb")

	require.NoError(t, os.WriteFile(video, nil, 0o600))
	require.NoError(t, os.WriteFile(thumb, nil, 0o600))

	files, err := walk([]string{dir}, nil, nil)
	require.NoError(t, err)
	require.Equal(t, []*File{{File: video, Thumb: thumb}}, files)
}
