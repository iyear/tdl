package up

import (
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/iyear/tdl/pkg/consts"
	"github.com/iyear/tdl/pkg/utils"
)

func walk(paths, excludes []string) ([]*file, error) {
	files := make([]*file, 0)
	excludesMap := map[string]struct{}{
		consts.UploadThumbExt: {}, // ignore thumbnail files
	}

	for _, exclude := range excludes {
		excludesMap[exclude] = struct{}{}
	}

	for _, path := range paths {
		err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if _, ok := excludesMap[filepath.Ext(path)]; ok {
				return nil
			}

			f := file{file: path}
			t := strings.TrimRight(path, filepath.Ext(path)) + consts.UploadThumbExt
			if utils.FS.PathExists(t) {
				f.thumb = t
			}

			files = append(files, &f)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return files, nil
}
