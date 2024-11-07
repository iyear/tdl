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
	LogPath = filepath.Join(DataDir, "log")
	ExtensionsPath = filepath.Join(DataDir, "extensions")
	ExtensionsDataPath = filepath.Join(ExtensionsPath, "data")

	for _, p := range []string{DataDir, ExtensionsPath, ExtensionsDataPath} {
		if err = os.MkdirAll(p, 0o755); err != nil {
			panic(err)
		}
	}
}
