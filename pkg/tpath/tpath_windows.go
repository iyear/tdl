//go:build windows

package tpath

import (
	"os"
	"path/filepath"
)

// https://github.com/telegramdesktop/tdesktop/blob/dev/Telegram/SourceFiles/platform/win/specific_win.cpp#L237-L249
func desktopAppData(_ string) []string {
	paths := make([]string, 0)
	dataDir := os.Getenv("APPDATA")
	if dataDir == "" {
		return paths
	}

	paths = append(paths,
		filepath.Join(dataDir, AppName),
		filepath.Join(dataDir, "Telegram Desktop UWP"))

	return paths
}
