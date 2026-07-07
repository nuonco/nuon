package ociarchive

import (
	"fmt"
	"os"
)

func (a *archive) TmpDir() string {
	return a.tmpDir
}

func (a *archive) createTmpDir() (string, error) {
	tmpDir, err := os.MkdirTemp("", "archive")
	if err != nil {
		return "", fmt.Errorf("unable to create temp dir: %s", err)
	}

	return tmpDir, nil
}
