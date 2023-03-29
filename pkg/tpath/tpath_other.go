//go:build !linux && !darwin && !windows

package tpath

func desktopAppData(_ string) []string {
	return []string{}
}
