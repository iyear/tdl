package downloader

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/gotd/contrib/partio"
)

type fakeSource struct{ data []byte }

func (s fakeSource) Chunk(_ context.Context, offset int64, b []byte) (int64, error) {
	if offset >= int64(len(s.data)) {
		return 0, io.EOF
	}
	return int64(copy(b, s.data[offset:])), nil
}

// TestStreamAtResumesPartialFile locks in the resume integration: seeking the
// output to a non-chunk-aligned partial offset and streaming the remainder must
// reproduce the full file without corrupting or duplicating the existing prefix.
func TestStreamAtResumesPartialFile(t *testing.T) {
	const (
		chunk   = 4
		partial = 6 // deliberately not a multiple of chunk
	)
	data := []byte("ABCDEFGHIJ")

	path := filepath.Join(t.TempDir(), "f.tmp")
	if err := os.WriteFile(path, data[:partial], 0o644); err != nil {
		t.Fatal(err)
	}

	f, err := os.OpenFile(path, os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	offset, err := f.Seek(0, io.SeekEnd)
	if err != nil {
		t.Fatal(err)
	}
	if offset != partial {
		t.Fatalf("precondition: offset = %d, want %d", offset, partial)
	}

	if err := partio.NewStreamer(fakeSource{data: data}, chunk).
		StreamAt(context.Background(), offset, f); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(data) {
		t.Fatalf("resumed content = %q, want %q", got, data)
	}
}
