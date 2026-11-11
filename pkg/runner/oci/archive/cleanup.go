package ociarchive

import (
	"context"
	"errors"
	"os"

	"github.com/spf13/afero"
)

func (a *archive) Cleanup(ctx context.Context) error {
	var errs []error
	if a.store != nil {
		errs = append(errs, a.store.Close())
	}
	if a.workFS != nil {
		entries, err := afero.ReadDir(a.workFS, ".")
		errs = append(errs, err)
		for _, entry := range entries {
			errs = append(errs, a.workFS.RemoveAll(entry.Name()))
		}
		errs = append(errs, a.workFS.Close())
		a.workFS = nil
	}
	if a.tmpDir != "" {
		// Only remove an empty directory at the display path, never recursively
		// delete a replacement tree if the job root was renamed.
		if err := os.Remove(a.tmpDir); err != nil && !os.IsNotExist(err) {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
