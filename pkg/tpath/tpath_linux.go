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

	path := make([]string, 0)

	prefix := filepath.Join(homedir, ".local", "share")
	path = append(path,
		filepath.Join(prefix, AppName),
		// https://github.com/iyear/tdl/issues/92#issuecomment-1699307412
		filepath.Join(prefix, "KotatogramDesktop"),
		filepath.Join(prefix, "64Gram"),
		filepath.Join(prefix, "TelegramDesktop"),
	)

	if t, err := filepath.Glob("~/snap/telegram-desktop/*/.local/share/TelegramDesktop"); err == nil {
		path = append(path, t...)
	}

	return path
}
