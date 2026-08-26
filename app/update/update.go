package update

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/Masterminds/semver/v3"
	"github.com/fatih/color"
	"github.com/go-faster/errors"
	"github.com/google/go-github/v62/github"
	"github.com/spf13/viper"
	"golang.org/x/net/proxy"

	"github.com/iyear/tdl/core/util/netutil"
	"github.com/iyear/tdl/pkg/consts"
)

type Options struct {
	Yes    bool
	DryRun bool   // report availability without downloading anything
	Target string // install a specific release tag instead of the latest (e.g. v0.20.4)
	Force  bool   // reinstall even when up to date; also allows downgrades
}

// Run performs a self-update: checks the latest GitHub release, downloads the
// matching archive, verifies its checksum and atomically replaces the current
// binary.
func Run(ctx context.Context, opts Options) (rerr error) {
	exe, err := os.Executable()
	if err != nil {
		return errors.Wrap(err, "get executable path")
	}
	if exe, err = filepath.EvalSymlinks(exe); err != nil {
		return errors.Wrap(err, "resolve executable path")
	}

	if strings.Contains(filepath.ToSlash(exe), "/Cellar/") {
		color.Yellow("tdl seems to be installed via Homebrew, please update it with 'brew upgrade tdl' instead.")
		if !opts.Yes {
			return nil
		}
		color.Yellow("--yes is set, updating anyway...")
	}

	dialer, err := netutil.NewProxy(viper.GetString(consts.FlagProxy))
	if err != nil {
		dialer = proxy.Direct
	}

	release, err := fetchRelease(ctx, opts.Target, dialer)
	if err != nil {
		return errors.Wrap(err, "fetch release")
	}

	tag := release.GetTagName()

	fmt.Printf("%s: %s\n", color.BlueString("Current version"), color.CyanString(consts.Version))
	if opts.Target != "" {
		fmt.Printf("%s:  %s\n", color.BlueString("Target version"), color.CyanString(tag))
	} else {
		fmt.Printf("%s:  %s\n", color.BlueString("Latest version"), color.CyanString(tag))
	}

	switch needsUpdate(consts.Version, tag) {
	case updateNo:
		if !opts.Force {
			color.Green("You are already using the latest version.")
			return nil
		}
		color.Yellow("--force is set, reinstalling %s.", tag)
	case updateUnknown:
		color.Yellow("Unrecognized current version (%s), will update to %s.", consts.Version, tag)
	default:
		// updateYes: proceed with the update below.
	}

	if opts.DryRun {
		color.Green("An update to %s is available.", tag)
		return nil
	}

	if !opts.Yes {
		ok := false

		if err = survey.AskOne(&survey.Confirm{
			Message: fmt.Sprintf("Update to %s?", tag),
		}, &ok); err != nil {
			return errors.Wrap(err, "confirm (use --yes to skip the prompt)")
		}

		if !ok {
			color.Red("Aborted.")
			return nil
		}
	}

	goarm := goARM()
	name, ok := assetName(runtime.GOOS, runtime.GOARCH, goarm)
	if !ok {
		return errors.Errorf("no release assets for platform %s/%s/%s", runtime.GOOS, runtime.GOARCH, goarm)
	}

	asset, err := findAsset(release, name)
	if err != nil {
		return err
	}
	sumsAsset, err := findAsset(release, checksumAssetName)
	if err != nil {
		return err
	}

	tmp, err := os.MkdirTemp("", "tdl-update-")
	if err != nil {
		return errors.Wrap(err, "create temp dir")
	}
	defer func() { rerr = multiErr(rerr, os.RemoveAll(tmp)) }()

	color.Cyan("Downloading %s...", asset.GetName())
	archivePath, _, err := download(ctx, asset.GetBrowserDownloadURL(), filepath.Join(tmp, name), dialer)
	if err != nil {
		return errors.Wrap(err, "download archive")
	}

	color.Cyan("Verifying checksum...")
	sumsPath, _, err := download(ctx, sumsAsset.GetBrowserDownloadURL(), filepath.Join(tmp, checksumAssetName), dialer)
	if err != nil {
		return errors.Wrap(err, "download checksums")
	}
	sumsData, err := os.ReadFile(sumsPath)
	if err != nil {
		return errors.Wrap(err, "read checksums")
	}
	if err = verifyChecksum(sumsData, filepath.Base(archivePath), archivePath); err != nil {
		return errors.Wrap(err, "verify checksum")
	}

	color.Cyan("Extracting %s...", name)
	binPath, err := extractBinary(archivePath, binaryName(runtime.GOOS), tmp)
	if err != nil {
		return errors.Wrap(err, "extract binary")
	}

	color.Cyan("Replacing %s...", exe)
	if err = replaceBinary(exe, binPath); err != nil {
		return errors.Wrap(err, "replace binary")
	}

	color.Green("Successfully updated to %s. Enjoy!", tag)
	return nil
}

// updateState is the result of comparing the current version with the latest one.
type updateState int

const (
	updateYes updateState = iota
	updateNo
	updateUnknown
)

// needsUpdate compares two version strings ("v1.2.3" or "1.2.3"). It returns
// updateUnknown if the current version is not a parseable release version
// (e.g. "dev"), so that dev builds can still self-update.
func needsUpdate(current, latest string) updateState {
	cur, err := semver.NewVersion(strings.TrimPrefix(current, "v"))
	if err != nil {
		return updateUnknown
	}

	lat, err := semver.NewVersion(strings.TrimPrefix(latest, "v"))
	if err != nil {
		return updateUnknown
	}

	if cur.Compare(lat) >= 0 {
		return updateNo
	}

	return updateYes
}

// fetchRelease returns the latest release, or the release tagged target if
// target is not empty (e.g. "v0.20.4").
func fetchRelease(ctx context.Context, target string, dialer proxy.ContextDialer) (*github.RepositoryRelease, error) {
	client := github.NewClient(&http.Client{
		Transport: &http.Transport{
			DialContext: dialer.DialContext,
		},
		Timeout: 30 * time.Second,
	})

	// Use GITHUB_TOKEN if set, to avoid hitting the unauthenticated rate limit.
	if ghToken := os.Getenv("GITHUB_TOKEN"); ghToken != "" {
		client = client.WithAuthToken(ghToken)
	}

	var (
		release *github.RepositoryRelease
		err     error
	)

	if target == "" {
		release, _, err = client.Repositories.GetLatestRelease(ctx, repoOwner, repoName)
	} else {
		if _, err = semver.NewVersion(strings.TrimPrefix(target, "v")); err != nil {
			return nil, fmt.Errorf("invalid target version %q: %w", target, err)
		}
		release, _, err = client.Repositories.GetReleaseByTag(ctx, repoOwner, repoName, target)
	}
	if err != nil {
		return nil, errors.Wrap(err, "github api")
	}

	if release == nil || release.GetTagName() == "" {
		return nil, fmt.Errorf("release not found")
	}

	return release, nil
}

func findAsset(release *github.RepositoryRelease, name string) (*github.ReleaseAsset, error) {
	for _, a := range release.Assets {
		if a.GetName() == name {
			return a, nil
		}
	}
	return nil, fmt.Errorf("asset %q not found in release %s", name, release.GetTagName())
}

// goARM returns the GOARM value used to build the binary ("7" as fallback),
// or an empty string on non-arm platforms.
func goARM() string {
	if runtime.GOARCH != "arm" {
		return ""
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, s := range info.Settings {
			if s.Key == "GOARM" && s.Value != "" {
				return s.Value
			}
		}
	}
	return "7"
}

func download(ctx context.Context, url, path string, dialer proxy.ContextDialer) (string, int64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", 0, errors.Wrap(err, "create request")
	}

	resp, err := (&http.Client{
		Transport: &http.Transport{
			DialContext: dialer.DialContext,
		},
		Timeout: 10 * time.Minute,
	}).Do(req)
	if err != nil {
		return "", 0, errors.Wrap(err, "do request")
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	f, err := os.Create(path)
	if err != nil {
		return "", 0, errors.Wrap(err, "create file")
	}
	defer func() { _ = f.Close() }()

	size, err := io.Copy(f, resp.Body)
	if err != nil {
		return "", 0, errors.Wrap(err, "save file")
	}

	return path, size, nil
}

func verifyChecksum(sumsContent []byte, name, path string) error {
	expected, ok := parseChecksums(string(sumsContent))[name]
	if !ok {
		return fmt.Errorf("checksum for %q not found", name)
	}

	f, err := os.Open(path)
	if err != nil {
		return errors.Wrap(err, "open file")
	}
	defer func() { _ = f.Close() }()

	h := sha256.New()
	if _, err = io.Copy(h, f); err != nil {
		return errors.Wrap(err, "hash file")
	}

	got := hex.EncodeToString(h.Sum(nil))
	if got != expected {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expected, got)
	}
	return nil
}

// extractBinary extracts the executable from a .tar.gz or .zip archive into dir.
func extractBinary(archivePath, binName, dir string) (string, error) {
	dest := filepath.Join(dir, binName)

	switch {
	case strings.HasSuffix(archivePath, ".zip"):
		if err := extractFromZip(archivePath, binName, dest); err != nil {
			return "", err
		}
	case strings.HasSuffix(archivePath, ".tar.gz"):
		if err := extractFromTarGz(archivePath, binName, dest); err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("unsupported archive format: %s", archivePath)
	}

	// 0755: the binary must be executable for everyone, like the install script does.
	if err := os.Chmod(dest, 0o755); err != nil {
		return "", errors.Wrap(err, "chmod binary")
	}
	return dest, nil
}

func extractFromTarGz(archivePath, binName, dest string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return errors.Wrap(err, "open archive")
	}
	defer func() { _ = f.Close() }()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return errors.Wrap(err, "gzip reader")
	}
	defer func() { _ = gz.Close() }()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return fmt.Errorf("%q not found in archive", binName)
		}
		if err != nil {
			return errors.Wrap(err, "tar next")
		}
		if filepath.Base(hdr.Name) != binName || hdr.Typeflag != tar.TypeReg {
			continue
		}
		out, err := os.OpenFile(dest, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
		if err != nil {
			return errors.Wrap(err, "create output file")
		}
		if _, err = io.Copy(out, tr); err != nil {
			_ = out.Close()
			return errors.Wrap(err, "copy binary")
		}
		return out.Close()
	}
}

func extractFromZip(archivePath, binName, dest string) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return errors.Wrap(err, "zip open reader")
	}
	defer func() { _ = r.Close() }()

	for _, zf := range r.File {
		if filepath.Base(zf.Name) != binName || zf.FileInfo().IsDir() {
			continue
		}
		src, err := zf.Open()
		if err != nil {
			return errors.Wrap(err, "open zip entry")
		}
		out, err := os.OpenFile(dest, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
		if err != nil {
			_ = src.Close()
			return errors.Wrap(err, "create output file")
		}
		if _, err = io.Copy(out, src); err != nil {
			_ = out.Close()
			_ = src.Close()
			return errors.Wrap(err, "copy binary")
		}
		if err = out.Close(); err != nil {
			_ = src.Close()
			return errors.Wrap(err, "close output file")
		}
		return src.Close()
	}
	return fmt.Errorf("%q not found in archive", binName)
}

// replaceBinary moves the new binary over the current executable. On unix it
// is an atomic rename within the same filesystem; on windows the old binary
// is kept as "<name>.old" until a successful replacement for rollback.
func replaceBinary(target, src string) error {
	dir := filepath.Dir(target)
	staged := filepath.Join(dir, "."+filepath.Base(target)+".update")

	if err := copyFile(src, staged, 0o755); err != nil {
		if errors.Is(err, os.ErrPermission) {
			return errors.Wrapf(err, "stage new binary in %s: permission denied (check directory ownership, or use your package manager)", dir)
		}
		return errors.Wrap(err, "stage new binary")
	}

	if runtime.GOOS == osWindows {
		// a running .exe cannot be deleted or renamed over, but it can be
		// renamed away: move the old binary to "<name>.old", put the new one
		// in place, and roll back from the backup if anything fails.
		bak := target + ".old"
		_ = os.Remove(bak)
		if err := os.Rename(target, bak); err != nil && !errors.Is(err, os.ErrNotExist) {
			_ = os.Remove(staged)
			return errors.Wrap(err, "backup old binary")
		}
		if err := os.Rename(staged, target); err != nil {
			_ = os.Rename(bak, target) // rollback
			return errors.Wrap(err, "replace binary")
		}
		_ = os.Remove(bak) // best-effort cleanup of the backup
		return nil
	}

	if err := os.Rename(staged, target); err != nil {
		_ = os.Remove(staged)
		return errors.Wrap(err, "replace binary")
	}
	return nil
}

func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return errors.Wrap(err, "open source")
	}
	defer func() { _ = in.Close() }()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return errors.Wrap(err, "create destination")
	}
	if _, err = io.Copy(out, in); err != nil {
		_ = out.Close()
		return errors.Wrap(err, "copy content")
	}
	if err = out.Sync(); err != nil {
		_ = out.Close()
		return errors.Wrap(err, "sync destination")
	}
	return out.Close()
}

func multiErr(a, b error) error {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	return fmt.Errorf("%v; %v", a, b)
}
