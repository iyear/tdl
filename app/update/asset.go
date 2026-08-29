package update

import (
	"fmt"
	"strings"
)

const (
	repoOwner = "iyear"
	repoName  = "tdl"

	checksumAssetName = repoName + "_checksums.txt"

	osWindows = "windows"
	osLinux   = "linux"
	osDarwin  = "darwin"
)

// goosName maps runtime.GOOS to the OS part of the release asset name,
// following the goreleaser naming scheme (see .goreleaser.yaml).
func goosName(goos string) (string, bool) {
	switch goos {
	case osLinux:
		return "Linux", true
	case osDarwin:
		return "MacOS", true
	case osWindows:
		return "Windows", true
	default:
		return "", false
	}
}

// archName maps GOARCH (and GOARM for arm builds) to the arch part of the
// release asset name.
func archName(goarch, goarm string) (string, bool) {
	switch goarch {
	case "386":
		return "32bit", true
	case "amd64":
		return "64bit", true
	case "arm":
		switch goarm {
		case "5", "6", "7":
			return "armv" + goarm, true
		default:
			return "", false
		}
	case "arm64", "riscv64", "loong64":
		return goarch, true
	default:
		return "", false
	}
}

// assetName returns the archive asset filename for the given platform,
// e.g. tdl_Linux_64bit.tar.gz or tdl_Windows_64bit.zip. The second return
// value is false if the platform has no release assets.
func assetName(goos, goarch, goarm string) (string, bool) {
	osn, ok := goosName(goos)
	if !ok {
		return "", false
	}

	arch, ok := archName(goarch, goarm)
	if !ok {
		return "", false
	}

	ext := ".tar.gz"
	if goos == osWindows {
		ext = ".zip"
	}

	return fmt.Sprintf("%s_%s_%s%s", repoName, osn, arch, ext), true
}

// binaryName returns the executable file name inside the release archive.
func binaryName(goos string) string {
	if goos == osWindows {
		return repoName + ".exe"
	}
	return repoName
}

// parseChecksums parses a goreleaser checksums.txt file
// ("sha256sum" format: "<hex>  <filename>") into a map keyed by filename.
// Lines whose digest is not a valid sha256 hex digest are ignored.
func parseChecksums(content string) map[string]string {
	sums := make(map[string]string)
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) != 2 || !isSHA256Hex(parts[0]) {
			continue
		}
		name := strings.TrimPrefix(parts[1], "*")
		sums[name] = strings.ToLower(parts[0])
	}
	return sums
}

// isSHA256Hex reports whether s is a 64-character lowercase/uppercase hex string.
func isSHA256Hex(s string) bool {
	if len(s) != sha256Len {
		return false
	}
	for _, c := range s {
		isDigit := c >= '0' && c <= '9'
		isLower := c >= 'a' && c <= 'f'
		isUpper := c >= 'A' && c <= 'F'
		if !isDigit && !isLower && !isUpper {
			return false
		}
	}
	return true
}

const sha256Len = 64
