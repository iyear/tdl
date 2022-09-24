package archive

import (
	"context"
	"github.com/fatih/color"
	"github.com/iyear/tdl/pkg/consts"
	"github.com/mholt/archiver/v4"
	"os"
)

func Backup(ctx context.Context, dst string) error {
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	files, err := archiver.FilesFromDisk(nil, map[string]string{
		consts.DataDir: "",
	})
	if err != nil {
		return err
	}

	format := archiver.Zip{}
	if err = format.Archive(ctx, f, files); err != nil {
		return err
	}

	color.Green("Backup successfully, file: %s", dst)
	return nil
}
