package up

import (
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveThumbUsesFileSize(t *testing.T) {
	path := filepath.Join(t.TempDir(), "video.thumb")

	f, err := os.Create(path)
	require.NoError(t, err)
	require.NoError(t, jpeg.Encode(f, image.NewRGBA(image.Rect(0, 0, 1, 1)), nil))
	require.NoError(t, f.Close())

	stat, err := os.Stat(path)
	require.NoError(t, err)

	thumb, err := (&iter{}).resolveThumb(path)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, thumb.Close()) })

	require.Equal(t, stat.Size(), thumb.Size())
}
