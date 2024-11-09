package internal

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
)

func FindFilesInChildDirs(regex *regexp.Regexp) ([]string, error) {
	filePaths := make([]string, 0)

	e := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if filepath.Dir(path) == "." {
			return nil
		}
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

func FindFileInCurrentDir(regex *regexp.Regexp) (string, error) {
	entries, err := os.ReadDir("./")
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		if regex.MatchString(entry.Name()) {
			return entry.Name(), nil
		}
	}
	return "", errors.New("no file found")
}
