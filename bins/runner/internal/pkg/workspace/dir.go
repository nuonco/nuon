package workspace

import (
	"io/fs"
	"os"
	"path/filepath"
)

const (
	defaultWorkspaceDirPermissions fs.FileMode = 0o777
)

func (w *workspace) rootDir() string {
	return filepath.Join(w.TmpRootDir, w.ID)
}

func (w *workspace) initRootDir() error {
	w.L.Info("initializing new workspace at " + w.rootDir())
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

func (w *workspace) Root() string {
	return w.rootDir()
}

func (w *workspace) AbsPath(path string) string {
	return filepath.Join(w.rootDir(), path)
}

func (w *workspace) IsFile(path string) bool {
	fp := w.AbsPath(path)
	stat, err := os.Stat(fp)
	if err != nil {
		return false
	}
	if stat.IsDir() {
		return false
	}

	return true
}

func (w *workspace) IsDir(path string) bool {
	fp := w.AbsPath(path)
	stat, err := os.Stat(fp)
	if err != nil {
		return false
	}
	return stat.IsDir()
}

func (w *workspace) RmDir(path string) error {
	fp := w.AbsPath(path)
	if !w.IsDir(path) {
		return nil
	}

	if err := os.RemoveAll(fp); err != nil {
		return err
	}

	return nil
}

func (w *workspace) IsExecutable(path string) bool {
	fp := w.AbsPath(path)
	stat, err := os.Stat(fp)
	if err != nil {
		return false
	}
	if stat.IsDir() {
		return false
	}

	return stat.Mode()&0o111 != 0
}
