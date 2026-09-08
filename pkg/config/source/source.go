package source

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/mitchellh/go-homedir"
)

// TODO(jm): this should be a global decoder hook
func ReadSource(val string) ([]byte, error) {
	return ReadSourceFrom(val, "")
}

func ReadSourceFrom(val, baseDir string) ([]byte, error) {
	path, err := expandSourcePath(val, baseDir)
	if err != nil {
		return nil, fmt.Errorf("unable to expand source path: %w", err)
	}

	byts, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read file: %w", err)
	}

	return byts, nil
}

func LoadSource(val string) (map[string]interface{}, error) {
	return LoadSourceFrom(val, "")
}

func LoadSourceFrom(val, baseDir string) (map[string]interface{}, error) {
	byts, err := ReadSourceFrom(val, baseDir)
	if err != nil {
		return nil, err
	}

	var obj map[string]interface{}
	if err := toml.Unmarshal(byts, &obj); err != nil {
		return nil, fmt.Errorf("unable to parse toml: %w", err)
	}

	return obj, nil
}

func expandSourcePath(source, baseDir string) (string, error) {
	path, err := homedir.Expand(source)
	if err != nil {
		return "", fmt.Errorf("unable to expand directory")
	}
	if baseDir != "" && !filepath.IsAbs(path) {
		path = filepath.Join(baseDir, path)
	}
	path, err = filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("unable to expand path")
	}

	_, err = os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("source %s does not exist", source)
		}
	}

	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		return "", fmt.Errorf("unable to evaluate symlinks on path")
	}

	return path, nil
}
