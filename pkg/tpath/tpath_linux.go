//go:build linux

package tpath

import (
	"github.com/iyear/tdl/pkg/consts"
	"github.com/iyear/tdl/pkg/utils"
	"path/filepath"
)

// https://github.com/telegramdesktop/tdesktop/blob/dev/Telegram/SourceFiles/platform/linux/specific_linux.cpp#L669-L684
func desktopAppData() []string {
	oldPath := filepath.Join(consts.HomeDir, ".TelegramDesktop")
	suffixes := []string{"0", "1", "s"}
	for _, s := range suffixes {
		if utils.FS.PathExists(filepath.Join(oldPath, "tdata", "settings"+s)) {
			return []string{oldPath}
		}
	}

	return []string{filepath.Join(consts.HomeDir, ".local", "share", AppName)}
}
