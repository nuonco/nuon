package workspace

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"
)

// cleanup cleans up the root directory and all contents
func (w *workspace) Cleanup(ctx context.Context) error {
	if c, ok := w.Backend.(io.Closer); ok && c != nil {
		if err := c.Close(); err != nil {
			return errors.Wrap(err, "unable to close terraform backend")
		}
	}

	if w.DisableCleanup {
		return nil
	}

	if err := w.Binary.Uninstall(ctx); err != nil {
		return errors.Wrap(err, "unable to uninstall terraform")
	}

	if err := w.Archive.Cleanup(ctx); err != nil {
		return errors.Wrap(err, "unable to cleanup archive")
	}

	if err := os.RemoveAll(w.root); err != nil {
		return fmt.Errorf("unable to remove %s: %w", w.root, err)
	}

	return nil
}
