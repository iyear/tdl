package extensions

import (
	"context"
	"path/filepath"
	"strings"
)

//go:generate go-enum --values --names --flag --nocase

const Prefix = "tdl-"

// ExtensionType ENUM(github, local)
type ExtensionType string

type Extension interface {
	Type() ExtensionType
	Name() string // Extension Name without tdl- prefix
	Path() string // Path to executable
	URL() string
	Owner() string
	CurrentVersion() string
	LatestVersion(ctx context.Context) string
	UpdateAvailable(ctx context.Context) bool
}

type baseExtension struct {
	path string
}

func (e baseExtension) Name() string {
	s := strings.TrimPrefix(filepath.Base(e.path), Prefix)
	s = strings.TrimSuffix(s, filepath.Ext(s))
	return s
}

func (e baseExtension) Path() string {
	return e.path
}

type manifest struct {
	Owner string `json:"owner,omitempty"`
	Repo  string `json:"repo,omitempty"`
	Tag   string `json:"tag,omitempty"`
}
