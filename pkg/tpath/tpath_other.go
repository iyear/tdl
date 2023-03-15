//go:build !linux && !darwin && !windows

package tpath

func desktopAppData() []string {
	return []string{}
}
