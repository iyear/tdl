package up

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
	"golang.org/x/sync/errgroup"

	"github.com/iyear/tdl/core/uploader"
	"github.com/iyear/tdl/core/util/mediautil"
)

const (
	maxThumbSizeBytes = 200 * 1024
	maxThumbWidth     = 320
	maxThumbHeight    = 320
)

var albumPattern = regexp.MustCompile(`^(.+?)_(\d{4}-\d{2}-\d{2})_(\d{2}-\d{2}-\d{2})`)

type preparedItem struct {
	media    uploader.MediaDescriptor
	peer     peers.Peer
	thread   int
	entities []tg.MessageEntityClass
	size     int64
}

type videoMetadata struct {
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	Duration int    `json:"duration"`
	Rotation int    `json:"rotation"`
	Codec    string `json:"codec"`
	Bitrate  int    `json:"bitrate"`
}

type ffprobeOutput struct {
	Streams []struct {
		CodecType string `json:"codec_type"`
		CodecName string `json:"codec_name"`
		Width     int    `json:"width"`
		Height    int    `json:"height"`
		Duration  string `json:"duration"`
		BitRate   string `json:"bit_rate"`
		Tags      struct {
			Rotate string `json:"rotate"`
		} `json:"tags"`
		SideDataList []struct {
			Rotation int `json:"rotation"`
		} `json:"side_data_list"`
	} `json:"streams"`
	Format struct {
		Duration string `json:"duration"`
		BitRate  string `json:"bit_rate"`
	} `json:"format"`
}

func prepareUploadTasks(ctx context.Context, files []*File, resolver *taskResolver, opts Options, limit int) ([]*uploadTask, error) {
	items := make([]*preparedItem, len(files))
	wg, wgctx := errgroup.WithContext(ctx)
	wg.SetLimit(max(1, limit))

	for idx, f := range files {
		idx, f := idx, f
		wg.Go(func() error {
			peer, thread, caption, entities, err := resolver.resolve(wgctx, f)
			if err != nil {
				return err
			}

			mime, err := mimetype.DetectFile(f.File)
			if err != nil {
				return errors.Wrapf(err, "detect file mime: %s", f.File)
			}

			stat, err := os.Stat(f.File)
			if err != nil {
				return errors.Wrapf(err, "stat file: %s", f.File)
			}

			mediaType := decideMediaType(f.File, mime.String(), opts)
			media := uploader.MediaDescriptor{
				Path:      f.File,
				Thumb:     f.Thumb,
				Mime:      mime.String(),
				Streaming: false,
				Caption:   caption,
				AlbumID:   albumGroupID(f.File),
				Type:      mediaType,
			}

			if mediaType == uploader.MediaTypeVideo {
				hash, err := fileSHA256(f.File)
				if err != nil {
					return errors.Wrap(err, "hash video")
				}

				meta, err := loadOrExtractMetadata(f.File, hash)
				if err != nil {
					return errors.Wrapf(err, "extract metadata: %s", f.File)
				}
				media.Width = meta.Width
				media.Height = meta.Height
				media.Duration = meta.Duration
				media.Streaming = true

				thumb, err := loadOrGenerateThumb(f.File, f.Thumb, hash)
				if err != nil {
					return errors.Wrapf(err, "prepare thumbnail: %s", f.File)
				}
				media.Thumb = thumb
			}

			items[idx] = &preparedItem{
				media:    media,
				peer:     peer,
				thread:   thread,
				entities: entities,
				size:     stat.Size(),
			}
			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, err
	}

	return buildTasks(items, opts), nil
}

func decideMediaType(path, mime string, opts Options) uploader.MediaType {
	switch {
	case mediautil.IsImage(mime) && opts.Photo && mime != "image/webp":
		return uploader.MediaTypePhoto
	case opts.Video && (mediautil.IsVideo(mime) || isDetectVideoExt(path)):
		return uploader.MediaTypeVideo
	case opts.DetectVideo && isDetectVideoExt(path):
		return uploader.MediaTypeVideo
	default:
		return uploader.MediaTypeDocument
	}
}

func buildTasks(items []*preparedItem, opts Options) []*uploadTask {
	if !opts.Album {
		ret := make([]*uploadTask, 0, len(items))
		for _, item := range items {
			label := fmt.Sprintf("%s -> %s(%d)", item.media.Path, item.peer.VisibleName(), item.peer.ID())
			ret = append(ret, &uploadTask{
				media:      []uploader.MediaDescriptor{item.media},
				to:         item.peer,
				thread:     item.thread,
				remove:     opts.Remove,
				paths:      []string{item.media.Path},
				totalSize:  item.size,
				label:      label,
				entityByID: map[string][]tg.MessageEntityClass{item.media.Path: item.entities},
			})
		}
		return ret
	}

	grouped := map[string][]*preparedItem{}
	order := make([]string, 0)
	for _, item := range items {
		album := item.media.AlbumID
		if album == "" {
			album = item.media.Path
		}
		k := fmt.Sprintf("%s|%d|%d", album, item.peer.ID(), item.thread)
		if _, ok := grouped[k]; !ok {
			order = append(order, k)
		}
		grouped[k] = append(grouped[k], item)
	}

	ret := make([]*uploadTask, 0, len(items))
	for _, k := range order {
		g := grouped[k]
		sort.Slice(g, func(i, j int) bool {
			return g[i].media.Path < g[j].media.Path
		})

		for start := 0; start < len(g); start += 10 {
			end := min(start+10, len(g))
			chunk := g[start:end]

			media := make([]uploader.MediaDescriptor, 0, len(chunk))
			paths := make([]string, 0, len(chunk))
			entityByID := make(map[string][]tg.MessageEntityClass, len(chunk))
			totalSize := int64(0)

			for i, item := range chunk {
				m := item.media
				entities := item.entities
				if len(chunk) > 1 && i > 0 {
					m.Caption = ""
					entities = nil
				}
				media = append(media, m)
				paths = append(paths, m.Path)
				totalSize += item.size
				entityByID[m.Path] = entities
			}

			ret = append(ret, &uploadTask{
				media:      media,
				to:         chunk[0].peer,
				thread:     chunk[0].thread,
				remove:     opts.Remove,
				paths:      paths,
				totalSize:  totalSize,
				label:      fmt.Sprintf("%s (%d itens) -> %s(%d)", media[0].AlbumID, len(media), chunk[0].peer.VisibleName(), chunk[0].peer.ID()),
				entityByID: entityByID,
			})
		}
	}

	return ret
}

func albumGroupID(path string) string {
	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	m := albumPattern.FindStringSubmatch(name)
	if len(m) != 4 {
		return ""
	}
	return fmt.Sprintf("%s_%s", m[1], m[2])
}

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", errors.Wrap(err, "open file")
	}
	defer func() { _ = f.Close() }()

	h := sha256.New()
	if _, err = io.Copy(h, f); err != nil {
		return "", errors.Wrap(err, "hash file")
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func cacheDirs() (thumbs string, metadata string, err error) {
	base := ".cache"
	thumbs = filepath.Join(base, "thumbs")
	metadata = filepath.Join(base, "metadata")
	if err = os.MkdirAll(thumbs, 0o755); err != nil {
		return "", "", err
	}
	if err = os.MkdirAll(metadata, 0o755); err != nil {
		return "", "", err
	}
	return thumbs, metadata, nil
}

func loadOrExtractMetadata(path, hash string) (videoMetadata, error) {
	_, metadataDir, err := cacheDirs()
	if err != nil {
		return videoMetadata{}, errors.Wrap(err, "prepare cache dir")
	}

	cachePath := filepath.Join(metadataDir, hash+".json")
	if data, readErr := os.ReadFile(cachePath); readErr == nil {
		var meta videoMetadata
		if jsonErr := json.Unmarshal(data, &meta); jsonErr == nil {
			return meta, nil
		}
	}

	meta, err := extractVideoMetadata(path)
	if err != nil {
		return videoMetadata{}, err
	}

	raw, err := json.Marshal(meta)
	if err == nil {
		_ = os.WriteFile(cachePath, raw, 0o644)
	}
	return meta, nil
}

func extractVideoMetadata(path string) (videoMetadata, error) {
	out, err := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_streams", "-show_format", path).Output()
	if err != nil {
		return videoMetadata{}, errors.Wrap(err, "run ffprobe")
	}

	var parsed ffprobeOutput
	if err = json.Unmarshal(out, &parsed); err != nil {
		return videoMetadata{}, errors.Wrap(err, "decode ffprobe output")
	}

	meta := videoMetadata{}
	for _, s := range parsed.Streams {
		if s.CodecType != "video" {
			continue
		}
		meta.Width = s.Width
		meta.Height = s.Height
		meta.Duration = parseDurationSeconds(s.Duration, parsed.Format.Duration)
		meta.Codec = s.CodecName
		meta.Bitrate = parseInt(s.BitRate, parseInt(parsed.Format.BitRate, 0))
		meta.Rotation = parseRotation(s.Tags.Rotate, s.SideDataList)
		break
	}

	if meta.Duration == 0 {
		meta.Duration = parseDurationSeconds("", parsed.Format.Duration)
	}

	return meta, nil
}

func parseRotation(tag string, sideData []struct {
	Rotation int `json:"rotation"`
}) int {
	if tag != "" {
		if v, err := strconv.Atoi(tag); err == nil {
			return v
		}
	}
	if len(sideData) > 0 {
		return sideData[0].Rotation
	}
	return 0
}

func parseDurationSeconds(primary, fallback string) int {
	toInt := func(v string) int {
		if v == "" {
			return 0
		}
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0
		}
		return int(f + 0.5)
	}
	if s := toInt(primary); s > 0 {
		return s
	}
	return toInt(fallback)
}

func parseInt(v string, def int) int {
	if v == "" {
		return def
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return def
	}
	return int(n)
}

func loadOrGenerateThumb(videoPath, explicitThumb, hash string) (string, error) {
	thumbsDir, _, err := cacheDirs()
	if err != nil {
		return "", errors.Wrap(err, "prepare cache dir")
	}

	cachePath := filepath.Join(thumbsDir, hash+".jpg")
	if ok, _ := validateThumb(cachePath); ok {
		return cachePath, nil
	}

	source := explicitThumb
	if source == "" {
		source = sidecarThumb(videoPath)
	}

	if source == "" {
		raw := filepath.Join(thumbsDir, hash+"_raw.jpg")
		_ = os.Remove(raw)
		if err = exec.Command("ffmpeg", "-y", "-ss", "00:00:02", "-i", videoPath, "-vframes", "1", "-vf", "scale=320:320:force_original_aspect_ratio=decrease", "-q:v", "2", raw).Run(); err != nil {
			return "", errors.Wrap(err, "generate thumbnail with ffmpeg")
		}
		source = raw
	}

	if err = normalizeThumb(source, cachePath); err != nil {
		return "", errors.Wrap(err, "normalize thumbnail")
	}

	return cachePath, nil
}

func sidecarThumb(videoPath string) string {
	base := strings.TrimSuffix(videoPath, filepath.Ext(videoPath))
	for _, ext := range []string{".jpg", ".jpeg", ".png"} {
		candidate := base + ext
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return ""
}

func normalizeThumb(inputPath, outputPath string) error {
	if filepath.Clean(inputPath) == filepath.Clean(outputPath) {
		if ok, err := validateThumb(outputPath); ok {
			return nil
		} else if err != nil {
			return err
		}
	}

	qualityCandidates := []int{2, 4, 8, 12, 16, 20, 24, 28, 31}
	for _, q := range qualityCandidates {
		cmd := exec.Command(
			"ffmpeg", "-y", "-i", inputPath,
			"-vframes", "1",
			"-vf", "scale=320:320:force_original_aspect_ratio=decrease",
			"-q:v", strconv.Itoa(q),
			"-f", "image2",
			outputPath,
		)
		if err := cmd.Run(); err != nil {
			continue
		}

		if ok, _ := validateThumb(outputPath); ok {
			return nil
		}
	}

	return errors.New("unable to create valid thumbnail <=200KB and <=320x320 JPEG")
}

func validateThumb(path string) (bool, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	if stat.Size() > maxThumbSizeBytes {
		return false, nil
	}

	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer func() { _ = f.Close() }()

	cfg, format, err := image.DecodeConfig(f)
	if err != nil {
		return false, err
	}
	if !slices.Contains([]string{"jpeg", "jpg"}, strings.ToLower(format)) {
		return false, nil
	}
	if cfg.Width > maxThumbWidth || cfg.Height > maxThumbHeight {
		return false, nil
	}
	return true, nil
}
