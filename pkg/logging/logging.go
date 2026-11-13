// Package logging configures the CLI's file logger.
package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/shirou/gopsutil/v3/process"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/iyear/tdl/core/util/logutil"
)

// Match only files owned by this logger, including Lumberjack's rotated backups.
var logName = regexp.MustCompile(`^tdl-([1-9][0-9]*)-[0-9]{8}T[0-9]{6}-[a-zA-Z0-9]+(?:-[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}-[0-9]{2}-[0-9]{2}\.[0-9]{3})?\.log(?:\.gz)?$`)

// New gives each invocation its own file and rotation history. A namespace is
// not sufficient: separate storage directories can use the same namespace.
func New(level zapcore.LevelEnabler, dir string) (*zap.Logger, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("create log directory: %w", err)
	}
	now := time.Now()
	cleanup(dir, now.Add(-7*24*time.Hour), process.PidExists)
	// CreateTemp reserves a unique name even if a PID is reused or multiple
	// loggers are initialized in one process within the same second.
	file, err := os.CreateTemp(dir, fmt.Sprintf("tdl-%d-%s-*.log", os.Getpid(), now.Format("20060102T150405")))
	if err != nil {
		return nil, fmt.Errorf("create log file: %w", err)
	}
	if err := file.Close(); err != nil {
		return nil, fmt.Errorf("close log file: %w", err)
	}
	return logutil.New(level, file.Name()), nil
}

// Cleanup is best effort. An old modification time does not imply that a
// process has exited: a transfer may run (or stay idle) for more than seven days.
// If a PID has been reused, retaining the old files is the conservative choice.
func cleanup(dir string, before time.Time, exists func(int32) (bool, error)) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		match := logName.FindStringSubmatch(entry.Name())
		if match == nil || !entry.Type().IsRegular() {
			continue
		}
		info, err := entry.Info()
		if err != nil || !info.ModTime().Before(before) {
			continue
		}
		pid, err := strconv.ParseInt(match[1], 10, 32)
		if err != nil {
			continue
		}
		alive, err := exists(int32(pid))
		if err != nil || alive {
			continue
		}
		// Another invocation may already have removed this file. Never recurse
		// into directories or follow symlinks, and leave legacy logs untouched.
		_ = os.Remove(filepath.Join(dir, entry.Name()))
	}
}
