package consts

import (
	"github.com/gotd/td/telegram"
	"runtime"
)

// TODO(iyear): usr -X flag to set id and hash
// Telegram desktop client
const (
	AppID   = 17349
	AppHash = "344583e45741c457fe1862106095a5eb"
)

var (
	Device = telegram.DeviceConfig{
		DeviceModel:   "tdl",
		SystemVersion: runtime.GOOS,
		AppVersion:    Version,
	}
)
