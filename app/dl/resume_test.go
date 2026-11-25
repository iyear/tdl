package dl

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenResumable(t *testing.T) {
	const total = 1000
	path := filepath.Join(t.TempDir(), "file.tmp")

	// fresh download: nothing on disk yet
	f, off, err := openResumable(path, total, false)
	if err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
	if off != 0 {
		t.Fatalf("fresh: offset = %d, want 0", off)
	}

	// partial file: resume from its size, keep existing bytes
	if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	f, off, err = openResumable(path, total, false)
	if err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
	if off != 5 {
		t.Fatalf("partial: offset = %d, want 5", off)
	}
	if b, _ := os.ReadFile(path); string(b) != "hello" {
		t.Fatalf("partial: content = %q, want %q", b, "hello")
	}

	// restart: truncate and start over
	f, off, err = openResumable(path, total, true)
	if err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
	if off != 0 {
		t.Fatalf("restart: offset = %d, want 0", off)
	}
	if b, _ := os.ReadFile(path); len(b) != 0 {
		t.Fatalf("restart: content len = %d, want 0", len(b))
	}

	// complete file: offset == size so the download is a no-op and gets renamed
	if err := os.WriteFile(path, make([]byte, total), 0o644); err != nil {
		t.Fatal(err)
	}
	f, off, err = openResumable(path, total, false)
	if err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
	if off != total {
		t.Fatalf("complete: offset = %d, want %d", off, total)
	}
}
