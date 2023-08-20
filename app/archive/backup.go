package archive

import (
	"context"
	"os"

	"github.com/fatih/color"
	"github.com/go-faster/errors"
	"github.com/mholt/archiver/v4"

	"github.com/iyear/tdl/pkg/consts"
)

func Backup(ctx context.Context, dst string) error {
	f, err := os.Create(dst)
	if err != nil {
		return errors.Wrap(err, "create file")
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	files, err := archiver.FilesFromDisk(nil, map[string]string{
		consts.KVPath: "",
	})
	if err != nil {
		return errors.Wrap(err, "walk files")
	}

	format := archiver.Zip{}
	if err = format.Archive(ctx, f, files); err != nil {
		return err
	}

	color.Green("Backup successfully, file: %s", dst)
	return nil
}
