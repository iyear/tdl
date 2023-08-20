// Package tdesktop exports functions from github.com/gotd/td/session/tdesktop package
package tdesktop

import (
	_ "unsafe" // for go:linkname

	_ "github.com/gotd/td/session/tdesktop" // for FileKey
)

//go:linkname FileKey github.com/gotd/td/session/tdesktop.fileKey
func FileKey(_ string) string
