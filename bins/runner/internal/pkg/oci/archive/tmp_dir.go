package ociarchive

import (
	"fmt"
	"io/ioutil"
)

func (a *archive) TmpDir() string {
	return a.tmpDir
} 

func (a *archive) createTmpDir() (string, error) {
	tmpDir, err := ioutil.TempDir("", "archive")
	if err != nil {
		return "", fmt.Errorf("unable to create temp dir: %s", err)
	}

	return tmpDir, nil
}
