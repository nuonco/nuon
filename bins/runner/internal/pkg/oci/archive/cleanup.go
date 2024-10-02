package ociarchive

import (
	"context"
	"fmt"
)

func (a *archive) Cleanup(ctx context.Context) error {
	if a.store == nil {
		return nil
	}

	if err := a.store.Close(); err != nil {
		return fmt.Errorf("unable to close file store backing archive: %w", err)
	}

	return nil
}
