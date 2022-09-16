package consts

import (
	"github.com/gotd/td/telegram"
	"runtime"
)

// https://core.telegram.org/api/obtaining_api_id#using-telegrams-open-source-code
// so we can not use Telegram Desktop ID
// below is iyear's application
const (
	AppID   = 15055931
	AppHash = "021d433426cbb920eeb95164498fe3d3"
)

var (
	Device = telegram.DeviceConfig{
		DeviceModel:   "tdl",
		SystemVersion: runtime.GOOS,
		AppVersion:    Version,
	}
)

const (
	ChatGroup   = "group"
	ChatPrivate = "private"
	ChatChannel = "channel"
	ChatUnknown = "unknown"
)
