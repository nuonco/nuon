package workspace

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// CleanupByID removes a single workspace directory by its ID.
// This is safe to call with parallel runner jobs since it only
// targets the specific workspace, not all workspace-prefixed dirs.
//
// It sweeps every root a workspace can be created under, since the caller is a
// generic per-job backstop that doesn't know which root the handler chose. A
// missing directory is not an error.
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
