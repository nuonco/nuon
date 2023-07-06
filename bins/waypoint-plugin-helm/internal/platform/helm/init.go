package helm

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/go-playground/validator/v10"
	"oras.land/oras-go/v2/content/file"
)

func New(v *validator.Validate) (*Platform, error) {
	tmpDir, err := ioutil.TempDir("", "helm-package-push")
	if err != nil {
		return nil, fmt.Errorf("unable to load temp dir: %s", err)
	}

	storeDir := filepath.Join(tmpDir, "store")
	store, err := file.New(storeDir)
	if err != nil {
		return nil, fmt.Errorf("unable to get file store: %w", err)
	}

	return &Platform{
		v:      v,
		tmpDir: tmpDir,
		store:  store,
	}, nil
}
