package tclient

const (
	AppBuiltin = "builtin"
	AppDesktop = "desktop"
)

var Apps = map[string]struct {
	AppID   int
	AppHash string
}{
	// application created by iyear
	AppBuiltin: {AppID: 15055931, AppHash: "021d433426cbb920eeb95164498fe3d3"},
	// application created by tdesktop.
	// https://opentele.readthedocs.io/en/latest/documentation/authorization/api/#class-telegramdesktop
	AppDesktop: {AppID: 2040, AppHash: "b18441a1ff607e10a989891a5462e627"},
}
