package services

import (
	"fmt"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
)

func rootDir() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", fmt.Errorf("unable to get home directory: %w", err)
	}

	return filepath.Join(home, "nuon/mono"), nil
}

func serviceDir(name string) (string, error) {
	root, err := rootDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(root, fmt.Sprintf("services/%s", name)), nil
}
