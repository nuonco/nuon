package ociarchive

import (
	"context"
	"fmt"
	"os"
)

func (a *archive) Cleanup(ctx context.Context) error {
	if a.tmpDir != "" {
		os.RemoveAll(a.tmpDir)
	}

	if a.store != nil {
		if err := a.store.Close(); err != nil {
			return fmt.Errorf("unable to close file store backing archive: %w", err)
		}
	}

	if a.basePath != "" {
		os.RemoveAll(a.basePath)
	}

	return nil
}
