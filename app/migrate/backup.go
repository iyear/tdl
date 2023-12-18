package migrate

import (
	"context"
	"encoding/json"
	"os"

	"github.com/fatih/color"
	"github.com/go-faster/errors"
	"github.com/klauspost/compress/zstd"
	"go.uber.org/multierr"

	"github.com/iyear/tdl/pkg/kv"
)

func Backup(ctx context.Context, dst string) (rerr error) {
	meta, err := kv.From(ctx).MigrateTo()
	if err != nil {
		return errors.Wrap(err, "read metadata")
	}

	f, err := os.Create(dst)
	if err != nil {
		return errors.Wrap(err, "create file")
	}
	defer multierr.AppendInvoke(&rerr, multierr.Close(f))

	enc, err := zstd.NewWriter(f, zstd.WithEncoderLevel(zstd.SpeedBestCompression))
	if err != nil {
		return errors.Wrap(err, "create zstd encoder")
	}
	defer multierr.AppendInvoke(&rerr, multierr.Close(enc))

	metaB, err := json.Marshal(meta)
	if err != nil {
		return errors.Wrap(err, "marshal metadata")
	}

	if _, err = enc.Write(metaB); err != nil {
		return errors.Wrap(err, "write metadata")
	}

	color.Green("Backup successfully, file: %s", dst)
	return nil
}
