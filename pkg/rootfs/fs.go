// Package rootfs provides an Afero filesystem pinned to an OS directory handle.
package rootfs

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/spf13/afero"
)

// FS owns a directory handle. Root is a display path, never an authority for I/O.
// Names must be relative. Symlink targets are preserved verbatim, but absolute
// and escaping links cannot be followed, even when they point back into Root.
type FS struct {
	root *os.Root
	path string
	once sync.Once
	err  error
}

var _ afero.Fs = (*FS)(nil)
var _ afero.Symlinker = (*FS)(nil)

// Open pins an existing directory, rejecting a symlink at the root itself.
// Ancestors of path must be trusted; use Sub for untrusted descendants.
func Open(path string) (*FS, error) {
	if runtime.GOOS == "js" || runtime.GOOS == "plan9" {
		return nil, fmt.Errorf("rootfs: directory handles unsupported on %s", runtime.GOOS)
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	before, err := os.Lstat(abs)
	if err != nil {
		return nil, err
	}
	if !before.IsDir() {
		return nil, &fs.PathError{Op: "openroot", Path: path, Err: fs.ErrInvalid}
	}
	r, err := os.OpenRoot(abs)
	if err != nil {
		return nil, err
	}
	after, err := r.Stat(".")
	if err != nil || !os.SameFile(before, after) {
		r.Close()
		return nil, fmt.Errorf("rootfs: root changed while opening %q", path)
	}
	return &FS{root: r, path: abs}, nil
}

// Sub derives an independently owned handle through the pinned parent, not
// through its possibly replaced display path. Closing either handle does not
// close the other.
func (f *FS) Sub(name string) (*FS, error) {
	if err := local("sub", name); err != nil {
		return nil, err
	}
	r, err := f.root.OpenRoot(name)
	if err != nil {
		return nil, err
	}
	return &FS{root: r, path: filepath.Join(f.path, name)}, nil
}

func (f *FS) Root() string { return f.path }
func (f *FS) Name() string { return "rootfs" }
func (f *FS) Close() error {
	f.once.Do(func() { f.err = f.root.Close() })
	return f.err
}

func local(op, name string) error {
	if !filepath.IsLocal(name) {
		return &fs.PathError{Op: op, Path: name, Err: fs.ErrPermission}
	}
	return nil
}

func (f *FS) Open(name string) (afero.File, error) {
	return f.OpenFile(name, os.O_RDONLY, 0)
}

func (f *FS) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	if err := local("open", name); err != nil {
		return nil, err
	}
	file, err := f.root.OpenFile(name, flag, perm)
	if err != nil {
		return nil, err
	}
	return &rootedFile{File: file, name: name}, nil
}

type rootedFile struct {
	*os.File
	name string
}

func (f *rootedFile) Name() string { return f.name }

func (f *FS) Create(name string) (afero.File, error) {
	return f.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o666)
}

func (f *FS) Stat(name string) (os.FileInfo, error) {
	if err := local("stat", name); err != nil {
		return nil, err
	}
	return f.root.Stat(name)
}

func (f *FS) LstatIfPossible(name string) (os.FileInfo, bool, error) {
	if err := local("lstat", name); err != nil {
		return nil, true, err
	}
	info, err := f.root.Lstat(name)
	return info, true, err
}

func (f *FS) ReadlinkIfPossible(name string) (string, error) {
	if err := local("readlink", name); err != nil {
		return "", err
	}
	return f.root.Readlink(name)
}

func (f *FS) SymlinkIfPossible(target, name string) error {
	if err := local("symlink", name); err != nil {
		return err
	}
	return f.root.Symlink(target, name)
}

func (f *FS) Mkdir(name string, perm os.FileMode) error {
	if err := local("mkdir", name); err != nil {
		return err
	}
	return f.root.Mkdir(name, perm)
}

func (f *FS) MkdirAll(name string, perm os.FileMode) error {
	if err := local("mkdirall", name); err != nil {
		return err
	}
	return f.root.MkdirAll(name, perm)
}

func (f *FS) Remove(name string) error {
	if err := local("remove", name); err != nil {
		return err
	}
	return f.root.Remove(name)
}

func (f *FS) RemoveAll(name string) error {
	if err := local("removeall", name); err != nil {
		return err
	}
	return f.root.RemoveAll(name)
}

func (f *FS) Rename(oldname, newname string) error {
	if err := local("rename", oldname); err != nil {
		return err
	}
	if err := local("rename", newname); err != nil {
		return err
	}
	return f.root.Rename(oldname, newname)
}

func (f *FS) Chmod(name string, mode os.FileMode) error {
	if err := local("chmod", name); err != nil {
		return err
	}
	return f.root.Chmod(name, mode)
}

func (f *FS) Chown(name string, uid, gid int) error {
	if err := local("chown", name); err != nil {
		return err
	}
	return f.root.Chown(name, uid, gid)
}

func (f *FS) Chtimes(name string, atime, mtime time.Time) error {
	if err := local("chtimes", name); err != nil {
		return err
	}
	return f.root.Chtimes(name, atime, mtime)
}
