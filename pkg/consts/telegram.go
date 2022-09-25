package consts

import (
	"github.com/gotd/td/telegram"
	"runtime"
)

const (
	AppBuiltin = "builtin"
	AppDesktop = "desktop"
)

var Apps = map[string]struct {
	AppID   int
	AppHash string
}{
	// application created by iyear
	AppBuiltin: {AppID: 15055931, AppHash: "021d433426cbb920eeb95164498fe3d3"},
	// application created by tdesktop.
	// https://github.com/telegramdesktop/tdesktop/blob/dev/Telegram/SourceFiles/config.h#L94-L95
	AppDesktop: {AppID: 17349, AppHash: "344583e45741c457fe1862106095a5eb"},
}

var (
	Device = telegram.DeviceConfig{
		DeviceModel:   "tdl",
		SystemVersion: runtime.GOOS,
		AppVersion:    Version,
	}
)

// External designation, different from Telegram mtproto
const (
	ChatGroup   = "group"
	ChatPrivate = "private"
	ChatChannel = "channel"
	ChatUnknown = "unknown"
)
