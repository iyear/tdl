package dl

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

const testStreamJSON = `{"dcId":1,"location":{"_":"inputDocumentFileLocation","id":"5046488299653301873","access_hash":"6151466871329478284","file_reference":[1,2,3]},"size":209528276,"mimeType":"video/mp4","fileName":"video.mp4"}`

func TestParseStreamPayload(t *testing.T) {
	t.Run("encoded stream path", func(t *testing.T) {
		payload, err := parseStreamPayload("stream/" + url.PathEscape(testStreamJSON))
		require.NoError(t, err)
		require.Equal(t, 1, payload.DCID)
		require.Equal(t, "5046488299653301873", payload.Location.ID)
		require.Equal(t, "6151466871329478284", payload.Location.AccessHash)
		require.Equal(t, []byte{1, 2, 3}, payload.Location.FileReference)
		require.Equal(t, int64(209528276), payload.Size)
		require.Equal(t, "video.mp4", payload.FileName)
	})

	t.Run("reject copied video tag", func(t *testing.T) {
		input := `<video playsinline="true" src="stream/` + url.PathEscape(testStreamJSON) + `"></video>`
		_, err := parseStreamPayload(input)
		require.ErrorContains(t, err, `stream payload must start with "stream/"`)
	})
}

func TestParseStreamPayloadValidation(t *testing.T) {
	_, err := parseStreamPayload("stream/" + url.PathEscape(`{"dcId":1,"location":{"_":"inputPhotoFileLocation"},"size":1}`))
	require.ErrorContains(t, err, `unsupported stream location`)
}
