package workspace

import (
	"io/fs"
	"os"
	"path/filepath"
)

const (
	defaultWorkspaceDirPermissions fs.FileMode = 0o600
)

func (w *workspace) rootDir() string {
	return filepath.Join(w.TmpRootDir, w.ID)
}

func (w *workspace) initRootDir() error {
	if err := os.MkdirAll(w.rootDir(), defaultWorkspaceDirPermissions); err != nil {
		return err
	}

	return nil
}

func (w *workspace) cleanupDir() error {
	if err := os.RemoveAll(w.rootDir()); err != nil {
		return err
	}

	return nil
}
