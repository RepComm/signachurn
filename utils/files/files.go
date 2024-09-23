package files

import (
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
)

func FileExists(reponame string) (bool, error) {
	if _, err := os.Stat(reponame); err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, err
	}
}

func FindFileExts(DirName string, Recurse bool, Exts []string) []string {
	results := []string{}

	filepath.Walk(DirName, func(fpath string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		ext := path.Ext(info.Name())
		if slices.Contains(Exts, ext) {
			results = append(results, fpath)
		}
		return nil
	})

	return results
}
