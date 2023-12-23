package consts

import (
	"os"
	"path/filepath"
)

func init() {
	dir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	HomeDir = dir
	DataDir = filepath.Join(dir, ".tdl")

	if err = os.MkdirAll(DataDir, os.ModePerm); err != nil {
		panic(err)
	}

	LogPath = filepath.Join(DataDir, "log")
}
