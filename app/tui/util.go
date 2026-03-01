package tui

import (
	"os/exec"
	"runtime"

	"github.com/gen2brain/beeep"
	"github.com/spf13/viper"
)

// openFile opens a file or URL in the default application
func openFile(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
		args = []string{url}
	}

	c := exec.Command(cmd, args...)
	return c.Start()
}

// notify sends a desktop notification
func notify(title, message string) error {
	if !viper.GetBool("notify") {
		return nil
	}
	// On Windows, appIcon can be empty or path to .ico
	return beeep.Notify(title, message, "")
}
