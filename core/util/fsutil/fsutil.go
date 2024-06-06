package fsutil

import (
	"os"
	"path/filepath"
	"strings"
)

func GetNameWithoutExt(path string) string {
	return strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

// AddPrefixDot add prefix dot if extension don't have
func AddPrefixDot(ext string) string {
	if !strings.HasPrefix(ext, ".") {
		return "." + ext
	}
	return ext
}
