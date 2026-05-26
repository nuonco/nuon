package workspace

import (
	"fmt"
	"os"
	"path/filepath"
)

// CleanupByID removes a single workspace directory by its ID.
// This is safe to call with parallel runner jobs since it only
// targets the specific workspace, not all workspace-prefixed dirs.
func CleanupByID(workspaceID string) error {
	dirPath := filepath.Join(defaultTmpRootDir, "workspace-"+workspaceID)
	if err := os.RemoveAll(dirPath); err != nil {
		return fmt.Errorf("failed to remove workspace directory %s: %w", dirPath, err)
	}
	return nil
}
