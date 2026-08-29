package update

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

// test constants to keep goconst happy
const (
	tLinux   = "linux"
	tDarwin  = "darwin"
	tWindows = "windows"
	tAmd64   = "amd64"
	tArm     = "arm"
	tArm64   = "arm64"

	vLatest = "v0.20.4"
)

func TestAssetName(t *testing.T) {
	tests := []struct {
		name   string
		goos   string
		goarch string
		goarm  string
		want   string
		ok     bool
	}{
		{"linux amd64", tLinux, tAmd64, "", "tdl_Linux_64bit.tar.gz", true},
		{"linux 386", tLinux, "386", "", "tdl_Linux_32bit.tar.gz", true},
		{"linux arm64", tLinux, tArm64, "", "tdl_Linux_arm64.tar.gz", true},
		{"linux riscv64", tLinux, "riscv64", "", "tdl_Linux_riscv64.tar.gz", true},
		{"linux loong64", tLinux, "loong64", "", "tdl_Linux_loong64.tar.gz", true},
		{"linux armv6", tLinux, tArm, "6", "tdl_Linux_armv6.tar.gz", true},
		{"linux armv7", tLinux, tArm, "7", "tdl_Linux_armv7.tar.gz", true},
		{"macos amd64", tDarwin, tAmd64, "", "tdl_MacOS_64bit.tar.gz", true},
		{"macos arm64", tDarwin, tArm64, "", "tdl_MacOS_arm64.tar.gz", true},
		{"windows amd64", tWindows, tAmd64, "", "tdl_Windows_64bit.zip", true},
		{"windows armv5", tWindows, tArm, "5", "tdl_Windows_armv5.zip", true},
		{"unsupported os", "freebsd", tAmd64, "", "", false},
		{"unsupported arch", tLinux, "mips", "", "", false},
		{"bad goarm", tLinux, tArm, "8", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := assetName(tt.goos, tt.goarch, tt.goarm)
			assert.Equal(t, tt.ok, ok)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestAssetNameAgainstRealRelease verifies that every platform built by
// goreleaser produces an asset name that matches a real release asset list.
func TestAssetNameAgainstRealRelease(t *testing.T) {
	releaseAssets := map[string]bool{
		"tdl_Linux_32bit.tar.gz": true, "tdl_Linux_64bit.tar.gz": true,
		"tdl_Linux_arm64.tar.gz": true, "tdl_Linux_armv5.tar.gz": true,
		"tdl_Linux_armv6.tar.gz": true, "tdl_Linux_armv7.tar.gz": true,
		"tdl_Linux_loong64.tar.gz": true, "tdl_Linux_riscv64.tar.gz": true,
		"tdl_MacOS_64bit.tar.gz": true, "tdl_MacOS_arm64.tar.gz": true,
		"tdl_Windows_32bit.zip": true, "tdl_Windows_64bit.zip": true,
		"tdl_Windows_arm64.zip": true, "tdl_Windows_armv5.zip": true,
		"tdl_Windows_armv6.zip": true, "tdl_Windows_armv7.zip": true,
	}

	for _, goarch := range []string{"386", tAmd64, tArm64, "riscv64", "loong64"} {
		got, ok := assetName(tLinux, goarch, "")
		assert.True(t, ok)
		assert.True(t, releaseAssets[got], "asset %s should exist in release", got)
	}
	for _, arm := range []string{"5", "6", "7"} {
		got, ok := assetName(tLinux, tArm, arm)
		assert.True(t, ok)
		assert.True(t, releaseAssets[got], "asset %s should exist in release", got)
	}
}

func TestBinaryName(t *testing.T) {
	assert.Equal(t, "tdl.exe", binaryName("windows"))
	assert.Equal(t, "tdl", binaryName(runtime.GOOS))
	if runtime.GOOS != "windows" {
		assert.Equal(t, "tdl", binaryName(runtime.GOOS))
	}
}

func TestParseChecksums(t *testing.T) {
	content := `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855  tdl_Linux_64bit.tar.gz
1111111111111111111111111111111111111111111111111111111111111111 *tdl_Windows_64bit.zip

invalid line
2222222222222222222222222222222222222222222222222222222222222222 tdl_MacOS_arm64.tar.gz
`
	sums := parseChecksums(content)

	assert.Len(t, sums, 3)
	assert.Equal(t, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", sums["tdl_Linux_64bit.tar.gz"])
	// binary-mode marker "*" must be stripped
	assert.Equal(t, "1111111111111111111111111111111111111111111111111111111111111111", sums["tdl_Windows_64bit.zip"])
	assert.Equal(t, "2222222222222222222222222222222222222222222222222222222222222222", sums["tdl_MacOS_arm64.tar.gz"])

	_, ok := sums["missing.tar.gz"]
	assert.False(t, ok)
}

func TestNeedsUpdate(t *testing.T) {
	tests := []struct {
		name    string
		current string
		latest  string
		want    updateState
	}{
		{"older current", "v0.20.3", vLatest, updateYes},
		{"same version", vLatest, vLatest, updateNo},
		{"no v prefix", "0.20.4", vLatest, updateNo},
		{"newer current", "v0.21.0", vLatest, updateNo},
		{"minor bump", "v0.19.9", "v0.20.0", updateYes},
		{"dev build", "dev", vLatest, updateUnknown},
		{"empty current", "", vLatest, updateUnknown},
		{"garbage current", "not-a-version", vLatest, updateUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, needsUpdate(tt.current, tt.latest))
		})
	}
}
