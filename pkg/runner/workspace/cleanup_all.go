package workspace

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func CleanupByID(workspaceID string) error {
	var errs []error
	for _, root := range HostActionRoots() {
		dirPath := filepath.Join(root, "workspace-"+workspaceID)
		if err := os.RemoveAll(dirPath); err != nil {
			errs = append(errs, fmt.Errorf("failed to remove workspace directory %s: %w", dirPath, err))
		}
	}
	return errors.Join(errs...)
}
