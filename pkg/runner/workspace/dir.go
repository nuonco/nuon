package workspace

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/nuonco/nuon/pkg/rootfs"
)

func (w *workspace) rootDir() string {
	if w.owned != nil {
		return w.owned.Root()
	}
	return filepath.Join(w.TmpRootDir, w.ID)
}

func (w *workspace) initRootDir() error {
	w.L.Info("initializing new workspace at " + w.rootDir())
	parent, child, err := rootfs.PrepareFS(w.backing, w.TmpRootDir, w.ID, false)
	if err != nil {
		return err
	}
	w.parent, w.fs = parent, child
	w.owned, _ = child.(*rootfs.FS)
	return nil
}

func (w *workspace) cleanupDir() error {
	if w.parent != nil {
		return w.parent.RemoveAll(w.ID)
	}
	if w.fs != nil {
		return w.fs.RemoveAll(".")
	}
	return rootfs.RemoveFS(w.backing, w.TmpRootDir, w.ID)
}

func (w *workspace) closeRoot() error {
	var errs []error
	if w.owned != nil {
		errs = append(errs, w.owned.Close())
		w.owned = nil
	}
	if w.parent != nil {
		errs = append(errs, w.parent.Close())
		w.parent = nil
	}
	w.fs = nil
	return errors.Join(errs...)
}

func (w *workspace) Root() string {
	return w.rootDir()
}

func (w *workspace) AbsPath(path string) string {
	return filepath.Join(w.rootDir(), path)
}

func (w *workspace) IsFile(path string) bool {
	if w.fs == nil {
		return false
	}
	stat, err := w.fs.Stat(relativePath(path))
	if err != nil {
		return false
	}
	if stat.IsDir() {
		return false
	}

	return true
}

func (w *workspace) IsDir(path string) bool {
	if w.fs == nil {
		return false
	}
	stat, err := w.fs.Stat(relativePath(path))
	if err != nil {
		return false
	}
	return stat.IsDir()
}

func (w *workspace) RmDir(path string) error {
	if !w.IsDir(path) {
		return nil
	}

	if err := w.fs.RemoveAll(relativePath(path)); err != nil {
		return err
	}

	return nil
}

func (w *workspace) IsExecutable(path string) bool {
	if w.fs == nil {
		return false
	}
	stat, err := w.fs.Stat(relativePath(path))
	if err != nil {
		return false
	}
	if stat.IsDir() {
		return false
	}

	return stat.Mode()&0o111 != 0
}

func relativePath(name string) string {
	name = strings.TrimLeft(name, "/")
	if name == "" {
		return "."
	}
	return name
}
