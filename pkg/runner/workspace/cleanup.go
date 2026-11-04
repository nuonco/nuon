package workspace

import (
	"context"
	"errors"
	"fmt"
	"os"
)

func (w *workspace) Cleanup(ctx context.Context) (retErr error) {
	defer func() { retErr = errors.Join(retErr, w.closeRoot()) }()
	if os.Getenv("IS_NUONCTL") == "true" {
		return nil
	}

	if err := w.cleanupDir(); err != nil {
		return fmt.Errorf("unable to cleanup directory: %w", err)
	}

	return nil
}
