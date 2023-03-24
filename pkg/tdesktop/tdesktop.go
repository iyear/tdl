// Package tdesktop exports functions from github.com/gotd/td/session/tdesktop package
package tdesktop

import (
	_ "github.com/gotd/td/session/tdesktop" // for FileKey
	_ "unsafe"                              // for go:linkname
)

//go:linkname FileKey github.com/gotd/td/session/tdesktop.fileKey
func FileKey(_ string) string
