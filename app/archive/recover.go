package archive

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/go-faster/errors"
	"github.com/mholt/archiver/v4"

	"github.com/iyear/tdl/pkg/consts"
)

func Recover(ctx context.Context, file string) error {
	f, err := os.Open(file)
	if err != nil {
		return errors.Wrap(err, "open file")
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	format := archiver.Zip{}

	if err = format.Extract(ctx, f, nil, func(ctx context.Context, af archiver.File) error {
		if af.IsDir() {
			return nil
		}

		v, err := af.Open()
		if err != nil {
			return errors.Wrap(err, "open archiver file")
		}
		defer func(v io.ReadCloser) {
			_ = v.Close()
		}(v)

		bytes, err := io.ReadAll(v)
		if err != nil {
			return errors.Wrap(err, "read all")
		}

		return os.WriteFile(filepath.Join(consts.DataDir, af.Name()), bytes, 0o644)
	}); err != nil {
		return err
	}

	color.Green("Recover successfully, file: %s", file)
	return nil
}
