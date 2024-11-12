package internal

import (
	"os"
	"path/filepath"
	"regexp"
)

func FindFilesRecursively(regex *regexp.Regexp) ([]string, error) {
	filePaths := make([]string, 0)

	e := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if err == nil && regex.MatchString(info.Name()) {
			filePaths = append(filePaths, path)
		}
		return nil
	})
	if e != nil {
		return filePaths, e
	}
	return filePaths, nil

}
