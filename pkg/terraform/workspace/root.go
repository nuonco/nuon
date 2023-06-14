package workspace

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

const (
	defaultDirPermissions      fs.FileMode = 0744
	defaultFilePermissions     fs.FileMode = 0600
	defaultFileExecPermissions fs.FileMode = 0777
)

func (w *workspace) Root() string {
	return w.root
}

// createRoot creates a new root directory for the workspace
func (w *workspace) createRoot() error {
	dir, err := os.MkdirTemp(w.tmpDirRoot, "workspace")
	if err != nil {
		return fmt.Errorf("unable to make temporary directory: %w", err)
	}
	w.root = dir
	return nil
}

// writeFile writes a file into the workspace
//
//nolint:unparam
func (w *workspace) writeFile(path string, byts []byte, perms fs.FileMode) error {
	fullPath := filepath.Join(w.root, path)
	dir := filepath.Dir(fullPath)

	if err := os.MkdirAll(dir, defaultDirPermissions); err != nil {
		return fmt.Errorf("unable to create root directory: %w", err)
	}

	if err := os.WriteFile(fullPath, byts, perms); err != nil {
		return fmt.Errorf("unable to write file: %w", err)
	}

	return nil
}

// cleanup cleans up the root directory and all contents
func (w *workspace) cleanup() error {
	if w.DisableCleanup {
		return nil
	}

	if err := os.RemoveAll(w.root); err != nil {
		return fmt.Errorf("unable to remove %s: %w", w.root, err)
	}

	return nil
}
