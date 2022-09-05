package consts

import (
	"github.com/gotd/td/telegram"
	"github.com/iyear/tdl/global"
	"os"
	"path/filepath"
	"runtime"
)

// TODO(iyear): usr -X flag to set id and hash
const (
	AppID   = 17349
	AppHash = "344583e45741c457fe1862106095a5eb"
)

const (
	DownloadModeURL = "url"
)

var (
	Device = telegram.DeviceConfig{
		DeviceModel:   "tdl",
		SystemVersion: runtime.GOOS,
		AppVersion:    global.Version,
	}
)

var (
	DataDir string
	KVPath  string
)

const (
	DocsPath     = "docs"
	DownloadPath = "downloads"
)

func init() {
	dir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	DataDir = filepath.Join(dir, ".tdl")

	if err = os.MkdirAll(DataDir, os.ModePerm); err != nil {
		panic(err)
	}

	KVPath = filepath.Join(DataDir, "data.kv")

	if err = os.MkdirAll(DownloadPath, os.ModePerm); err != nil {
		panic(err)
	}
}
