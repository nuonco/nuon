package workspace

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/powertoolsdev/mono/pkg/generics"
)

var cleanupPrefixes []string = []string{
	"workspace",
	"run",
	"plugin",
	"archive",
}

func CleanupAll(ctx context.Context) error {
	entries, err := os.ReadDir(defaultTmpRootDir)
	if err != nil {
		return fmt.Errorf("failed to read /tmp directory: %w", err)
	}

	for _, entry := range entries {
		// Check if context is done
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if !entry.IsDir() {
			continue
		}

		if generics.HasAnyPrefix(entry.Name(), cleanupPrefixes...) {
			dirPath := filepath.Join(defaultTmpRootDir, entry.Name())

			if err := os.RemoveAll(dirPath); err != nil {
				return fmt.Errorf("failed to remove workspace directory %s: %w", dirPath, err)
			}
		}
	}

	return nil
}
