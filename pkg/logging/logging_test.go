package logging

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Both writers open their files before either rotates. No Telegram session or
// terminal is involved: these are independent processes sharing one log directory.
func TestConcurrentRotation(t *testing.T) {
	dir := t.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	type child struct {
		cmd    *exec.Cmd
		in     io.WriteCloser
		out    *bufio.Scanner
		stderr bytes.Buffer
	}
	children := make([]*child, 2)
	for i := range children {
		c := &child{cmd: exec.CommandContext(ctx, os.Args[0], "-test.run=^TestLogWriterProcess$")}
		children[i] = c
		c.cmd.Env = append(os.Environ(), "TDL_LOG_TEST_DIR="+dir, fmt.Sprintf("TDL_LOG_TEST_ID=%d", i))
		c.cmd.Stderr = &c.stderr
		var err error
		c.in, err = c.cmd.StdinPipe()
		if err != nil {
			t.Fatal(err)
		}
		out, err := c.cmd.StdoutPipe()
		if err != nil {
			t.Fatal(err)
		}
		c.out = bufio.NewScanner(out)
		if err := c.cmd.Start(); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			_ = c.in.Close()
			_ = c.cmd.Process.Kill()
			_ = c.cmd.Wait()
		})
		if !c.out.Scan() || c.out.Text() != "ready" {
			t.Fatal("writer did not become ready")
		}
	}
	for _, c := range children {
		if _, err := fmt.Fprintln(c.in, "write"); err != nil {
			t.Fatal(err)
		}
		// Rotate one writer at a time while the other keeps its file open.
		// This makes the shared-file failure deterministic on Windows.
		if !c.out.Scan() || c.out.Text() != "written" {
			output := c.out.Text()
			for c.out.Scan() {
				output += "\n" + c.out.Text()
			}
			t.Fatalf("writer failed: %s", output)
		}
	}

	// Lumberjack compresses asynchronously. Keep the writers alive until both
	// archives are readable, and verify every record across current and backup files.
	deadline := time.Now().Add(10 * time.Second)
	for {
		if err := checkRecords(dir); err == nil {
			break
		} else if time.Now().After(deadline) {
			t.Fatal(err)
		}
		time.Sleep(20 * time.Millisecond)
	}
	for _, c := range children {
		if err := c.in.Close(); err != nil {
			t.Fatal(err)
		}
		if err := c.cmd.Wait(); err != nil {
			t.Fatalf("writer: %v: %s", err, c.stderr.String())
		}
		if c.stderr.Len() != 0 {
			t.Fatalf("unexpected stderr: %s", c.stderr.String())
		}
	}
}

func TestNew(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "logs")
	for i := 0; i < 2; i++ {
		if _, err := New(zap.InfoLevel, dir); err != nil {
			t.Fatal(err)
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("got %d files, want two distinct files", len(entries))
	}
	for _, entry := range entries {
		if !logName.MatchString(entry.Name()) {
			t.Fatalf("unrecognized filename: %s", entry.Name())
		}
	}
	// A non-directory path must produce an actionable startup error.
	if _, err := New(zap.InfoLevel, filepath.Join(dir, entries[0].Name())); err == nil {
		t.Fatal("expected error when log directory is a file")
	}
}

func TestCleanup(t *testing.T) {
	dir := t.TempDir()
	now := time.Now()
	old := now.Add(-8 * 24 * time.Hour)
	tests := []struct {
		name     string
		modified time.Time
		keep     bool
	}{
		{"tdl-1-20260101T000000-123.log", old, false},
		{"tdl-1-20260101T000000-123-2026-01-02T00-00-00.000.log", old, false},
		{"tdl-1-20260101T000000-123-2026-01-02T00-00-00.000.log.gz", old, false},
		{"tdl-1-20260101T000000-456.log", now, true},
		{"tdl-2-20260101T000000-123.log", old, true}, // live owner
		{"tdl-2-20260101T000000-123-2026-01-02T00-00-00.000.log.gz", old, true},
		{"tdl-3-20260101T000000-123.log", old, true}, // cannot inspect owner
		{"tdl-999999999999-20260101T000000-123.log", old, true},
		{"latest.log", old, true},
		{"latest-2026-01-02T00-00-00.000.log.gz", old, true},
		{"unrelated.log", old, true},
	}
	for _, tt := range tests {
		path := filepath.Join(dir, tt.name)
		if err := os.WriteFile(path, []byte("log"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(path, tt.modified, tt.modified); err != nil {
			t.Fatal(err)
		}
	}
	directory := filepath.Join(dir, "tdl-1-20260101T000000-directory.log")
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(directory, old, old); err != nil {
		t.Fatal(err)
	}
	exists := func(pid int32) (bool, error) {
		switch pid {
		case 1:
			return false, nil
		case 2:
			return true, nil
		default:
			return false, errors.New("access denied")
		}
	}
	// Repeating cleanup also covers files already removed by another invocation.
	for i := 0; i < 2; i++ {
		cleanup(dir, now.Add(-7*24*time.Hour), exists)
	}
	for _, tt := range tests {
		_, err := os.Stat(filepath.Join(dir, tt.name))
		if tt.keep && err != nil {
			t.Errorf("lost %s: %v", tt.name, err)
		}
		if !tt.keep && !os.IsNotExist(err) {
			t.Errorf("did not remove %s: %v", tt.name, err)
		}
	}
	if _, err := os.Stat(directory); err != nil {
		t.Fatalf("removed directory: %v", err)
	}
}

func TestNewPreservesLiveOwner(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, fmt.Sprintf("tdl-%d-20260101T000000-old.log", os.Getpid()))
	if err := os.WriteFile(path, []byte("old but active"), 0o600); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-8 * 24 * time.Hour)
	if err := os.Chtimes(path, old, old); err != nil {
		t.Fatal(err)
	}
	if _, err := New(zap.InfoLevel, dir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("removed live process log: %v", err)
	}
}

func TestLogWriterProcess(t *testing.T) {
	dir := os.Getenv("TDL_LOG_TEST_DIR")
	if dir == "" {
		return
	}
	logger, err := New(zap.InfoLevel, dir)
	if err != nil {
		t.Fatal(err)
	}
	var errors bytes.Buffer
	logger = logger.WithOptions(zap.ErrorOutput(zapcore.AddSync(&errors)))
	logger.Info("opened")
	fmt.Println("ready")
	in := bufio.NewScanner(os.Stdin)
	if !in.Scan() {
		return
	}
	payload := strings.Repeat("x", 1024*1024)
	for i := 0; i < 11; i++ {
		logger.Info(fmt.Sprintf("record-%s-%02d", os.Getenv("TDL_LOG_TEST_ID"), i), zap.String("payload", payload))
	}
	if err := logger.Sync(); err != nil {
		t.Fatal(err)
	}
	if errors.Len() != 0 {
		t.Fatalf("logging errors: %s", errors.String())
	}
	fmt.Println("written")
	in.Scan()
}

func checkRecords(dir string) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	counts := make(map[string]int)
	archives := 0
	for _, entry := range files {
		if strings.HasSuffix(entry.Name(), ".gz") {
			archives++
		}
	}
	// Opening an uncompressed backup on Windows can prevent Lumberjack from
	// deleting it after compression. Wait for the rename/compression work first.
	if archives != 2 || len(files) != 4 {
		return fmt.Errorf("got %d files and %d compressed backups, want 4 and 2", len(files), archives)
	}
	for _, entry := range files {
		name := entry.Name()
		if !strings.HasSuffix(name, ".log") && !strings.HasSuffix(name, ".log.gz") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return err
		}
		if strings.HasSuffix(name, ".gz") {
			reader, err := gzip.NewReader(bytes.NewReader(data))
			if err != nil {
				return err
			}
			data, err = io.ReadAll(reader)
			_ = reader.Close()
			if err != nil {
				return err
			}
		}
		for writer := 0; writer < 2; writer++ {
			for i := 0; i < 11; i++ {
				key := fmt.Sprintf("record-%d-%02d", writer, i)
				counts[key] += bytes.Count(data, []byte(key))
			}
		}
	}
	for key, count := range counts {
		if count != 1 {
			return fmt.Errorf("record %s occurs %d times, want 1", key, count)
		}
	}
	return nil
}
