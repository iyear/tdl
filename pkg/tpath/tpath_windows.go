//go:build windows

package tpath

import (
	"os"
	"path/filepath"
)

// https://github.com/telegramdesktop/tdesktop/blob/dev/Telegram/SourceFiles/platform/win/specific_win.cpp#L237-L249
func desktopAppData(_ string) (paths []string) {
	dataDir := os.Getenv("APPDATA")
	if dataDir == "" {
		return
	}

	paths = append(paths,
		filepath.Join(dataDir, AppName),
		filepath.Join(dataDir, "Telegram Desktop UWP"))

	return
}
