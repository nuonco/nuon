package parse

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

const (
	cfgFilePrefix string = "nuon."
	defaultFormat string = "toml"
)

var ErrInvalidFilename error = fmt.Errorf("invalid filename")

type File struct {
	AppName string
	Path    string
}

func FilenameFromAppName(appName string) string {
	return fmt.Sprintf("%s%s.%s", cfgFilePrefix, appName, defaultFormat)
}

func AppNameFromFilename(file string) (string, error) {
	pieces := strings.SplitN(file, ".", 3)
	if len(pieces) != 3 {
		return "", ErrInvalidFilename
	}
	appID := pieces[1]

	return appID, nil
}

func FindConfigFiles(rootDir string) ([]File, error) {
	cfgFiles := make([]File, 0)
	if err := filepath.WalkDir(rootDir, func(path string, _ fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if strings.HasPrefix(path, cfgFilePrefix) && strings.HasSuffix(path, defaultFormat) {
			appName, err := AppNameFromFilename(path)
			if err != nil {
				return err
			}

			cfgFiles = append(cfgFiles, File{
				AppName: appName,
				Path:    path,
			})
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("unable to look for current config files: %w", err)
	}

	return cfgFiles, nil
}
