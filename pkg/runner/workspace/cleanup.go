package workspace

import (
	"context"
	"fmt"
	"os"
)

func (w *workspace) Cleanup(ctx context.Context) error {
	if os.Getenv("IS_NUONCTL") == "true" {
		return nil
	}

	if err := w.cleanupDir(); err != nil {
		return fmt.Errorf("unable to cleanup directory: %w", err)
	}

	return nil
}
