//go:build darwin

package tpath

import (
	"github.com/iyear/tdl/pkg/consts"
	"path/filepath"
)

// https://github.com/telegramdesktop/tdesktop/blob/dev/Telegram/SourceFiles/platform/mac/specific_mac_p.mm#L364-L370
func desktopAppData() []string {
	return []string{filepath.Join(consts.HomeDir, "Library", "Application Support", AppName)}
}
