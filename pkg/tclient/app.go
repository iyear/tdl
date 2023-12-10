package tclient

import "github.com/gotd/td/telegram"

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
	// https://opentele.readthedocs.io/en/latest/documentation/authorization/api/#class-telegramdesktop
	AppDesktop: {AppID: 2040, AppHash: "b18441a1ff607e10a989891a5462e627"},
}

var Device = telegram.DeviceConfig{
	DeviceModel:    "Desktop",
	SystemVersion:  "Windows 10",
	AppVersion:     "4.2.4 x64",
	LangCode:       "en",
	SystemLangCode: "en-US",
	LangPack:       "tdesktop",
}
