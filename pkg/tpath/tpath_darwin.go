//go:build darwin

package tpath

import (
	"path/filepath"
)

// https://github.com/telegramdesktop/tdesktop/blob/dev/Telegram/SourceFiles/platform/mac/specific_mac_p.mm#L364-L370
func desktopAppData(homedir string) []string {
	return []string{filepath.Join(homedir, "Library", "Application Support", AppName)}
}
