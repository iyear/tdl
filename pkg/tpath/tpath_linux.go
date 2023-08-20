//go:build linux

package tpath

import (
	"path/filepath"

	"github.com/iyear/tdl/pkg/utils"
)

// https://github.com/telegramdesktop/tdesktop/blob/dev/Telegram/SourceFiles/platform/linux/specific_linux.cpp#L669-L684
func desktopAppData(homedir string) []string {
	oldPath := filepath.Join(homedir, ".TelegramDesktop")
	suffixes := []string{"0", "1", "s"}
	for _, s := range suffixes {
		if utils.FS.PathExists(filepath.Join(oldPath, "tdata", "settings"+s)) {
			return []string{oldPath}
		}
	}

	return []string{filepath.Join(homedir, ".local", "share", AppName)}
}
