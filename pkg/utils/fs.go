package utils

import (
	"path/filepath"
	"strings"
)

type fs struct{}

var FS = fs{}

func (f fs) GetNameWithoutExt(path string) string {
	return strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
}
