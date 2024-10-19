package extensions

import (
	"context"
	"fmt"
)

type localExtension struct {
	baseExtension
}

func (l *localExtension) Type() ExtensionType {
	return ExtensionTypeLocal
}

func (l *localExtension) URL() string {
	return fmt.Sprintf("file://%s", l.Path())
}

func (l *localExtension) Owner() string {
	return "local"
}

func (l *localExtension) CurrentVersion() string {
	return ""
}

func (l *localExtension) LatestVersion(_ context.Context) string {
	return ""
}

func (l *localExtension) UpdateAvailable(_ context.Context) bool {
	return false
}
