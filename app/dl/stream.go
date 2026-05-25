package dl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/gabriel-vasile/mimetype"
	"github.com/go-faster/errors"
	gotddownloader "github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/tg"
	"golang.org/x/sync/errgroup"

	"github.com/iyear/tdl/core/dcpool"
	coredownloader "github.com/iyear/tdl/core/downloader"
	"github.com/iyear/tdl/core/util/fsutil"
	"github.com/iyear/tdl/core/util/tutil"
	"github.com/iyear/tdl/pkg/utils"
)

type streamPayload struct {
	DCID     int                `json:"dcId"`
	Location streamFileLocation `json:"location"`
	Size     int64              `json:"size"`
	MIMEType string             `json:"mimeType"`
	FileName string             `json:"fileName"`
}

type streamFileLocation struct {
	Type          string `json:"_"`
	ID            string `json:"id"`
	AccessHash    string `json:"access_hash"`
	FileReference []byte `json:"file_reference"`
}

func downloadStreams(ctx context.Context, pool dcpool.Pool, opts Options, threads int, limit int) error {
	if err := os.MkdirAll(opts.Dir, 0o755); err != nil {
		return errors.Wrap(err, "create dir")
	}

	streams := make([]*streamPayload, 0, len(opts.Streams))
	for _, s := range opts.Streams {
		payload, err := parseStreamPayload(s)
		if err != nil {
			return errors.Wrap(err, "parse stream")
		}
		streams = append(streams, payload)
	}

	color.Green("All files will be downloaded to '%s' dir", opts.Dir)

	wg, wgctx := errgroup.WithContext(ctx)
	wg.SetLimit(limit)
	for _, payload := range streams {
		payload := payload
		wg.Go(func() error {
			return downloadStream(wgctx, pool, opts, threads, payload)
		})
	}
	return wg.Wait()
}

func downloadStream(ctx context.Context, pool dcpool.Pool, opts Options, threads int, payload *streamPayload) error {
	location, err := payload.Location.TG()
	if err != nil {
		return errors.Wrap(err, "stream location")
	}

	fileName := filepath.Base(strings.TrimSpace(payload.FileName))
	if fileName == "." || fileName == string(filepath.Separator) || fileName == "" {
		fileName = payload.Location.ID
		if ext := mimetype.Lookup(payload.MIMEType); ext != nil {
			fileName += ext.Extension()
		}
	}

	if len(opts.Include) > 0 && !hasExt(opts.Include, fileName) {
		return nil
	}
	if len(opts.Exclude) > 0 && hasExt(opts.Exclude, fileName) {
		return nil
	}

	finalPath := uniqueStreamPath(opts.Dir, fileName)
	if opts.SkipSame {
		if stat, err := os.Stat(finalPath); err == nil && stat.Size() == payload.Size {
			return nil
		}
	}

	tempPath := finalPath + tempExt
	tempFile, err := os.Create(tempPath)
	if err != nil {
		return errors.Wrap(err, "create file")
	}
	defer func() {
		_ = tempFile.Close()
		if _, err := os.Stat(tempPath); err == nil {
			_ = os.Remove(tempPath)
		}
	}()

	client := pool.Client(ctx, payload.DCID)
	if opts.Takeout {
		client = pool.Takeout(ctx, payload.DCID)
	}

	_, err = gotddownloader.NewDownloader().
		WithPartSize(coredownloader.MaxPartSize).
		Download(client, location).
		WithThreads(tutil.BestThreads(payload.Size, threads)).
		Parallel(ctx, tempFile)
	if err != nil {
		if strings.Contains(err.Error(), "FILE_REFERENCE_EXPIRED") {
			return errors.New("stream file reference expired; refresh the Telegram Web page and copy a fresh stream payload, or use the original message link")
		}
		return errors.Wrap(err, "download stream")
	}

	if err := tempFile.Close(); err != nil {
		return errors.Wrap(err, "close file")
	}

	newPath := finalPath
	if opts.RewriteExt {
		mime, err := mimetype.DetectFile(tempPath)
		if err != nil {
			return errors.Wrap(err, "detect mime")
		}
		if ext := mime.Extension(); ext != "" && filepath.Ext(newPath) != ext {
			newPath = filepath.Join(filepath.Dir(newPath), fsutil.GetNameWithoutExt(filepath.Base(newPath))+ext)
			newPath = uniqueStreamPath(filepath.Dir(newPath), filepath.Base(newPath))
		}
	}

	if err := os.Rename(tempPath, newPath); err != nil {
		return errors.Wrap(err, "rename file")
	}

	color.Green("Downloaded %s (%s)", newPath, utils.Byte.FormatBinaryBytes(payload.Size))
	return nil
}

func parseStreamPayload(input string) (*streamPayload, error) {
	raw := strings.TrimSpace(input)
	if !strings.HasPrefix(raw, "stream/") {
		return nil, errors.New(`stream payload must start with "stream/"`)
	}
	raw = strings.TrimPrefix(raw, "stream/")

	decoded, err := url.PathUnescape(raw)
	if err != nil {
		return nil, err
	}

	var payload streamPayload
	if err := json.Unmarshal([]byte(decoded), &payload); err != nil {
		return nil, err
	}
	if payload.Location.Type != "inputDocumentFileLocation" {
		return nil, fmt.Errorf("unsupported stream location %q", payload.Location.Type)
	}
	if payload.DCID == 0 {
		return nil, errors.New("missing dcId")
	}
	if payload.Size <= 0 {
		return nil, errors.New("missing size")
	}
	return &payload, nil
}

func (s streamFileLocation) TG() (*tg.InputDocumentFileLocation, error) {
	id, err := strconv.ParseInt(s.ID, 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, "parse id")
	}
	accessHash, err := strconv.ParseInt(s.AccessHash, 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, "parse access hash")
	}
	if len(s.FileReference) == 0 {
		return nil, errors.New("missing file reference")
	}
	return &tg.InputDocumentFileLocation{
		ID:            id,
		AccessHash:    accessHash,
		FileReference: s.FileReference,
	}, nil
}

func hasExt(exts []string, name string) bool {
	ext := filepath.Ext(name)
	for _, candidate := range exts {
		if fsutil.AddPrefixDot(candidate) == ext {
			return true
		}
	}
	return false
}

func uniqueStreamPath(dir string, name string) string {
	candidate := filepath.Join(dir, name)
	if _, err := os.Stat(candidate); os.IsNotExist(err) {
		return candidate
	}

	ext := filepath.Ext(name)
	stem := strings.TrimSuffix(name, ext)
	for i := 1; ; i++ {
		candidate = filepath.Join(dir, fmt.Sprintf("%s (%d)%s", stem, i, ext))
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
}
