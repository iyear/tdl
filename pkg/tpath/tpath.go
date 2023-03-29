package tpath

const AppName = "Telegram Desktop"

type desktop struct{}

var Desktop = desktop{}

// AppData returns possible paths of Telegram Desktop's data directory based on the current platform.
func (desktop) AppData(homedir string) []string {
	return desktopAppData(homedir)
}
