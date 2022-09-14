package up

import (
	"io/fs"
	"path/filepath"
)

func walk(paths, excludes []string) ([]string, error) {
	files := make([]string, 0)
	excludesMap := make(map[string]struct{})

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

			files = append(files, path)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return files, nil
}
