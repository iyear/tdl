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

	DataDir = filepath.Join(dir, ".tdl")

	if err = os.MkdirAll(DataDir, os.ModePerm); err != nil {
		panic(err)
	}

	KVPath = filepath.Join(DataDir, "data.kv")

	if err = os.MkdirAll(DownloadPath, os.ModePerm); err != nil {
		panic(err)
	}
}
