package internal

import (
	"os"
	"path/filepath"
	"regexp"
)

func FindFiles(regex string) ([]string, error) {
	filePaths := make([]string, 0)
	libRegEx, e := regexp.Compile(regex)
	if e != nil {
		return filePaths, e
	}

	e = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if err == nil && libRegEx.MatchString(info.Name()) {
			filePaths = append(filePaths, path)
		}
		return nil
	})
	if e != nil {
		return filePaths, e
	}
	return filePaths, nil

}
