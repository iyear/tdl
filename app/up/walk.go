package up

import (
	"github.com/iyear/tdl/pkg/filterMap"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/iyear/tdl/pkg/consts"
	"github.com/iyear/tdl/pkg/utils"
)

func walk(paths, includes, excludes []string) ([]*file, error) {
	files := make([]*file, 0)

	// include and exclude
	includesMap := filterMap.New(includes, utils.FS.AddPrefixDot)
	excludesMap := filterMap.New(excludes, utils.FS.AddPrefixDot)
	excludesMap[consts.UploadThumbExt] = struct{}{} // ignore thumbnail files

	for _, path := range paths {
		err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}

			// process include and exclude
			ext := filepath.Ext(path)
			if _, ok := includesMap[ext]; len(includesMap) > 0 && !ok {
				return nil
			}
			if _, ok := excludesMap[ext]; len(excludesMap) > 0 && ok {
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
