package extensions

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/go-faster/errors"
	"github.com/google/go-github/v62/github"
)

const (
	githubHost   = "github.com"
	manifestName = "manifest.json"
)

type githubExtension struct {
	baseExtension

	client *github.Client
	mu     sync.RWMutex

	// lazy loaded
	mf            *manifest
	latestVersion string
}

func (e *githubExtension) Type() ExtensionType {
	return ExtensionTypeGithub
}

func (e *githubExtension) URL() string {
	if mf, err := e.loadManifest(); err == nil {
		return fmt.Sprintf("https://%s/%s/%s", githubHost, mf.Owner, mf.Repo)
	}

	return ""
}

func (e *githubExtension) Owner() string {
	if mf, err := e.loadManifest(); err == nil {
		return mf.Owner
	}

	return ""
}

func (e *githubExtension) CurrentVersion() string {
	if mf, err := e.loadManifest(); err == nil {
		return mf.Tag
	}

	return ""
}

func (e *githubExtension) LatestVersion(ctx context.Context) string {
	e.mu.RLock()
	if e.latestVersion != "" {
		defer e.mu.RUnlock()
		return e.latestVersion
	}
	e.mu.RUnlock()

	mf, err := e.loadManifest()
	if err != nil {
		return ""
	}

	release, _, err := e.client.Repositories.GetLatestRelease(ctx, mf.Owner, mf.Repo)
	if err != nil {
		return ""
	}

	e.mu.Lock()
	e.latestVersion = release.GetTagName()
	e.mu.Unlock()

	return e.latestVersion
}

func (e *githubExtension) loadManifest() (*manifest, error) {
	e.mu.RLock()
	if e.mf != nil {
		defer e.mu.RUnlock()
		return e.mf, nil
	}
	e.mu.RUnlock()

	dir, _ := filepath.Split(e.Path())
	manifestPath := filepath.Join(dir, manifestName)

	var mfb []byte
	mfb, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, errors.Wrapf(err, "read manifest file %s", manifestPath)
	}

	mf := manifest{}
	if err = json.Unmarshal(mfb, &mf); err != nil {
		return nil, errors.Wrapf(err, "unmarshal manifest file %s", manifestPath)
	}

	e.mu.Lock()
	e.mf = &mf
	e.mu.Unlock()

	return e.mf, nil
}

func (e *githubExtension) UpdateAvailable(ctx context.Context) bool {
	if e.CurrentVersion() == "" ||
		e.LatestVersion(ctx) == "" ||
		e.CurrentVersion() == e.LatestVersion(ctx) {
		return false
	}
	return true
}
