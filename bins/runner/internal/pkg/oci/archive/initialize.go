package ociarchive

import (
	"context"
	"fmt"
	"path/filepath"

	"oras.land/oras-go/v2/content/file"
)

func (a *archive) Initialize(ctx context.Context) error {
	tmpDir, err := a.createTmpDir()
	if err != nil {
		return err
	}
	a.tmpDir = tmpDir

	storeDir := filepath.Join(tmpDir, "store")
	store, err := file.New(storeDir)
	if err != nil {
		return fmt.Errorf("unable to create file store: %w", err)
	}
	a.basePath = storeDir

	a.store = store
	return nil
}
