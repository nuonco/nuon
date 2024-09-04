package workspace

import (
	"context"
	"fmt"
)

func (w *workspace) Cleanup(ctx context.Context) error {
	if err := w.cleanupDir(); err != nil {
		return fmt.Errorf("unable to cleanup directory: %w", err)
	}

	return nil
}
