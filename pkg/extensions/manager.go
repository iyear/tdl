package extensions

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/go-faster/errors"
	"github.com/google/go-github/v62/github"
	"go.uber.org/multierr"

	"github.com/iyear/tdl/extension"
)

var (
	ErrAlreadyUpToDate = errors.New("already up to date")
	ErrOnlyGitHub      = errors.New("only GitHub extension can be upgraded by tdl")
)

type Manager struct {
	dir    string
	http   *http.Client
	github *github.Client

	dryRun bool
}

func NewManager(dir string) *Manager {
	return &Manager{
		dir:    dir,
		http:   http.DefaultClient,
		github: newGhClient(http.DefaultClient),
		dryRun: false,
	}
}

func newGhClient(c *http.Client) *github.Client {
	ghToken := os.Getenv("GITHUB_TOKEN")
	if ghToken == "" {
		return github.NewClient(c)
	}
	return github.NewClient(c).WithAuthToken(ghToken)
}

func (m *Manager) SetDryRun(v bool) {
	m.dryRun = v
}

func (m *Manager) DryRun() bool {
	return m.dryRun
}

func (m *Manager) SetClient(client *http.Client) {
	m.http = client
	m.github = newGhClient(client)
}

func (m *Manager) Dispatch(ext Extension, args []string, env *extension.Env, stdin io.Reader, stdout, stderr io.Writer) (rerr error) {
	cmd := exec.Command(ext.Path(), args...)

	envFile, err := os.CreateTemp("", "*")
	if err != nil {
		return errors.Wrap(err, "create temp")
	}
	defer func() { multierr.AppendInto(&rerr, os.Remove(envFile.Name())) }()

	envBytes, err := json.Marshal(env)
	if err != nil {
		return errors.Wrap(err, "marshal env")
	}

	if _, err = envFile.Write(envBytes); err != nil {
		return errors.Wrap(err, "write env to temp")
	}
	if err = envFile.Close(); err != nil {
		return errors.Wrap(err, "close env file")
	}

	cmd.Env = append(os.Environ(), fmt.Sprintf("%s=%s", extension.EnvKey, envFile.Name()))
	cmd.Args = append([]string{Prefix + ext.Name()}, args...) // reset args[0] to extension name instead of binary path
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	return cmd.Run()
}

func (m *Manager) List(ctx context.Context, includeLatestVersion bool) ([]Extension, error) {
	entries, err := os.ReadDir(m.dir)
	if err != nil {
		return nil, errors.Wrap(err, "read dir entries")
	}

	extensions := make([]Extension, 0, len(entries))
	for _, f := range entries {
		if !strings.HasPrefix(f.Name(), Prefix) {
			continue
		}

		if !f.IsDir() {
			continue
		}

		if _, err = os.Stat(filepath.Join(m.dir, f.Name(), manifestName)); err == nil {
			extensions = append(extensions, &githubExtension{
				baseExtension: baseExtension{path: filepath.Join(m.dir, f.Name(), f.Name())},
				client:        m.github,
			})
		} else {
			extensions = append(extensions, &localExtension{
				baseExtension: baseExtension{path: filepath.Join(m.dir, f.Name(), f.Name())},
			})
		}
	}

	if includeLatestVersion {
		m.populateLatestVersions(ctx, extensions)
	}

	return extensions, nil
}

// Upgrade only GitHub extension can be upgraded
func (m *Manager) Upgrade(ctx context.Context, ext Extension) error {
	switch e := ext.(type) {
	case *githubExtension:
		if !ext.UpdateAvailable(ctx) {
			return ErrAlreadyUpToDate
		}

		mf, err := e.loadManifest()
		if err != nil {
			return errors.Wrapf(err, "load manifest of %q", e.Name())
		}

		if !m.dryRun {
			if err = m.Remove(ext); err != nil {
				return errors.Wrapf(err, "remove old version extension")
			}
			if err = m.installGitHub(ctx, mf.Owner, mf.Repo, false); err != nil {
				return errors.Wrapf(err, "install GitHub extension %q", e.Name())
			}
		}

		return nil
	default:
		return ErrOnlyGitHub
	}
}

// Install installs an extension by target.
// Valid targets are:
// - GitHub: owner/repo
// - Local: path to executable.
func (m *Manager) Install(ctx context.Context, target string, force bool) error {
	// local
	if _, err := os.Stat(target); err == nil {
		return m.installLocal(target, force)
	}

	// github
	ownerRepo := strings.Split(target, "/")
	if len(ownerRepo) != 2 {
		return errors.Errorf("invalid target: %q", target)
	}

	return m.installGitHub(ctx, ownerRepo[0], ownerRepo[1], force)
}

func (m *Manager) installLocal(path string, force bool) error {
	src, err := os.Lstat(path)
	if err != nil {
		return errors.Wrap(err, "source extension stat")
	}
	if !src.Mode().IsRegular() {
		return errors.Errorf("invalid src extension: %q, only regular file is allowed", path)
	}

	name := src.Name()
	if !strings.HasPrefix(name, Prefix) {
		name = Prefix + name
	}

	targetDir := filepath.Join(m.dir, strings.TrimSuffix(name, filepath.Ext(name)))
	binPath := filepath.Join(targetDir, name)
	if err = m.maybeExist(binPath, force); err != nil {
		return err
	}

	if !m.dryRun {
		if err = os.MkdirAll(targetDir, 0o755); err != nil {
			return errors.Wrapf(err, "create target dir %q for extension %q", targetDir, name)
		}

		if err = copyRegularFile(path, binPath); err != nil {
			return errors.Wrapf(err, "install local extension: %q", path)
		}
	}

	return nil
}

func (m *Manager) installGitHub(ctx context.Context, owner, repo string, force bool) (rerr error) {
	if !strings.HasPrefix(repo, Prefix) {
		return errors.Errorf("invalid repo name: %q, should start with %q", repo, Prefix)
	}

	platform, ext := platformBinaryName()

	targetDir := filepath.Join(m.dir, repo)
	binPath := filepath.Join(targetDir, repo) + ext
	if err := m.maybeExist(binPath, force); err != nil {
		return err
	}

	release, _, err := m.github.Repositories.GetLatestRelease(ctx, owner, repo)
	if err != nil {
		return errors.Wrapf(err, "get latest release of %s/%s", owner, repo)
	}

	// match binary name
	var asset *github.ReleaseAsset
	for _, a := range release.Assets {
		if strings.HasSuffix(a.GetName(), platform+ext) {
			asset = a
			break
		}
	}

	if asset == nil {
		return errors.Errorf("no matched binary(%s) found in the release(%s)", platform+ext, release.GetHTMLURL())
	}

	if !m.dryRun {
		if err = os.MkdirAll(targetDir, 0o755); err != nil {
			return errors.Wrapf(err, "create target dir %q for extension %s/%s", targetDir, owner, repo)
		}

		if err = m.downloadGitHubAsset(ctx, owner, repo, asset, binPath); err != nil {
			return errors.Wrapf(err, "download github asset %s", asset.GetBrowserDownloadURL())
		}
	}

	mf := &manifest{
		Owner: owner,
		Repo:  repo,
		Tag:   release.GetTagName(),
	}

	mfb, err := json.Marshal(mf)
	if err != nil {
		return errors.Wrap(err, "marshal manifest")
	}

	if !m.dryRun {
		if err = os.WriteFile(filepath.Join(targetDir, manifestName), mfb, 0o644); err != nil {
			return errors.Wrapf(err, "write manifest to %s", targetDir)
		}
	}

	return nil
}

func (m *Manager) maybeExist(binPath string, force bool) error {
	targetDir := filepath.Dir(binPath)
	extName := filepath.Base(targetDir)

	if _, err := os.Lstat(binPath); err != nil {
		return nil
	}

	if !force {
		return errors.Errorf("extension already exists, please remove it first")
	}

	// force remove
	if !m.dryRun {
		if err := os.RemoveAll(targetDir); err != nil {
			return errors.Wrapf(err, "remove existing extension %q", extName)
		}
	}

	return nil
}

// Remove removes an extension by name(without prefix).
func (m *Manager) Remove(ext Extension) error {
	target := Prefix + ext.Name()
	targetDir := filepath.Join(m.dir, target)
	if _, err := os.Lstat(targetDir); os.IsNotExist(err) {
		return errors.Errorf("no extension found: %s", targetDir)
	}

	if !m.dryRun {
		return os.RemoveAll(targetDir)
	}

	return nil
}

func (m *Manager) populateLatestVersions(ctx context.Context, exts []Extension) {
	wg := &sync.WaitGroup{}
	for _, ext := range exts {
		wg.Add(1)
		go func(e Extension) {
			defer wg.Done()
			e.LatestVersion(ctx)
		}(ext)
	}
	wg.Wait()
}

func (m *Manager) downloadGitHubAsset(ctx context.Context, owner, repo string, asset *github.ReleaseAsset, dst string) (rerr error) {
	readCloser, _, err := m.github.Repositories.DownloadReleaseAsset(ctx, owner, repo, asset.GetID(), m.http)
	if err != nil {
		return errors.Wrapf(err, "download release asset %s", asset.GetName())
	}
	defer multierr.AppendInvoke(&rerr, multierr.Close(readCloser))

	file, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return errors.Wrapf(err, "open file %s", dst)
	}
	defer multierr.AppendInvoke(&rerr, multierr.Close(file))

	if _, err = io.Copy(file, readCloser); err != nil {
		return errors.Wrapf(err, "copy http body to %s", dst)
	}
	return nil
}

func copyRegularFile(src, dst string) (rerr error) {
	r, err := os.Open(src)
	if err != nil {
		return errors.Wrapf(err, "open src %s", src)
	}
	defer multierr.AppendInvoke(&rerr, multierr.Close(r))

	info, err := r.Stat()
	if err != nil {
		return errors.Wrapf(err, "stat file %s", src)
	}
	if !info.Mode().IsRegular() {
		return errors.Errorf("invalid source file: %q, only regular file is allowed", src)
	}

	w, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o666|info.Mode()&0o777)
	if err != nil {
		return errors.Wrapf(err, "open dst %s", dst)
	}
	defer multierr.AppendInvoke(&rerr, multierr.Close(w))

	if _, err = io.Copy(w, r); err != nil {
		return errors.Wrapf(err, "copy file %s to %s", src, dst)
	}
	return nil
}

func platformBinaryName() (string, string) {
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}

	arch := runtime.GOARCH
	switch arch {
	case "arm":
		if goarm := extractGOARM(); goarm != "" {
			arch += "v" + goarm
		}
	}

	return fmt.Sprintf("%s-%s", runtime.GOOS, arch), ext
}

func extractGOARM() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}

	for _, setting := range info.Settings {
		if setting.Key == "GOARM" {
			return setting.Value
		}
	}

	return ""
}
