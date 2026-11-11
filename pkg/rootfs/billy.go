package rootfs

import (
	"crypto/rand"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/go-git/go-billy/v5"
	"github.com/spf13/afero"
)

// Billy adapts a borrowed Afero filesystem for go-git. Close releases only the
// child handles opened by Chroot, not the borrowed filesystem.
type Billy struct {
	fs       afero.Fs
	path     string
	owned    *FS
	mu       sync.Mutex
	children []*Billy
	closed   bool
}

var _ billy.Filesystem = (*Billy)(nil)

func NewBilly(fs afero.Fs) *Billy {
	path := "/"
	if rooted, ok := fs.(*FS); ok {
		path = rooted.Root()
	}
	return &Billy{fs: fs, path: path}
}

func (b *Billy) Root() string               { return b.path }
func (b *Billy) Join(elem ...string) string { return filepath.Join(elem...) }

func (b *Billy) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return nil
	}
	b.closed = true
	var errs []error
	for _, child := range b.children {
		errs = append(errs, child.Close())
	}
	if b.owned != nil {
		errs = append(errs, b.owned.Close())
	}
	return errors.Join(errs...)
}

func (b *Billy) Chroot(name string) (billy.Filesystem, error) {
	if err := local("chroot", name); err != nil {
		return nil, err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return nil, os.ErrClosed
	}
	var child *Billy
	if rooted, ok := b.fs.(*FS); ok {
		sub, err := rooted.Sub(name)
		if err != nil {
			return nil, err
		}
		child = NewBilly(sub)
		child.owned = sub
	} else {
		child = NewBilly(afero.NewBasePathFs(b.fs, name))
		child.path = filepath.Join(b.path, name)
	}
	b.children = append(b.children, child)
	return child, nil
}

func (b *Billy) Open(name string) (billy.File, error) {
	return b.OpenFile(name, os.O_RDONLY, 0)
}

func (b *Billy) Create(name string) (billy.File, error) {
	return b.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o666)
}

func (b *Billy) OpenFile(name string, flag int, perm os.FileMode) (billy.File, error) {
	if err := local("open", name); err != nil {
		return nil, err
	}
	if flag&os.O_CREATE != 0 {
		if err := b.fs.MkdirAll(filepath.Dir(name), 0o755); err != nil {
			return nil, err
		}
	}
	f, err := b.fs.OpenFile(name, flag, perm)
	if err != nil {
		return nil, err
	}
	return &billyFile{File: f, name: name}, nil
}

func (b *Billy) Stat(name string) (os.FileInfo, error) {
	if err := local("stat", name); err != nil {
		return nil, err
	}
	return b.fs.Stat(name)
}

func (b *Billy) Lstat(name string) (os.FileInfo, error) {
	if err := local("lstat", name); err != nil {
		return nil, err
	}
	if fs, ok := b.fs.(afero.Lstater); ok {
		info, _, err := fs.LstatIfPossible(name)
		return info, err
	}
	return nil, billy.ErrNotSupported
}

func (b *Billy) Rename(oldname, newname string) error {
	if err := local("rename", oldname); err != nil {
		return err
	}
	if err := local("rename", newname); err != nil {
		return err
	}
	if err := b.MkdirAll(filepath.Dir(newname), 0o755); err != nil {
		return err
	}
	return b.fs.Rename(oldname, newname)
}

func (b *Billy) Remove(name string) error {
	if err := local("remove", name); err != nil {
		return err
	}
	return b.fs.Remove(name)
}

func (b *Billy) MkdirAll(name string, perm os.FileMode) error {
	if err := local("mkdirall", name); err != nil {
		return err
	}
	// go-git includes the directory type bit; os.Root accepts permissions only.
	return b.fs.MkdirAll(name, perm&^os.ModeDir)
}

func (b *Billy) ReadDir(name string) ([]os.FileInfo, error) {
	if name == "" {
		name = "."
	}
	if err := local("readdir", name); err != nil {
		return nil, err
	}
	return afero.ReadDir(b.fs, name)
}

func (b *Billy) TempFile(dir, prefix string) (billy.File, error) {
	if dir == "" {
		dir = "."
	}
	if err := local("tempfile", dir); err != nil {
		return nil, err
	}
	if err := b.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	if strings.ContainsAny(prefix, `/\`) {
		return nil, os.ErrInvalid
	}
	for range 10 {
		name := filepath.Join(dir, prefix+rand.Text())
		f, err := b.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o600)
		if os.IsExist(err) {
			continue
		}
		return f, err
	}
	return nil, os.ErrExist
}

func (b *Billy) Symlink(target, name string) error {
	if err := local("symlink", name); err != nil {
		return err
	}
	if fs, ok := b.fs.(afero.Linker); ok {
		if err := b.MkdirAll(filepath.Dir(name), 0o755); err != nil {
			return err
		}
		return fs.SymlinkIfPossible(target, name)
	}
	return billy.ErrNotSupported
}

func (b *Billy) Readlink(name string) (string, error) {
	if err := local("readlink", name); err != nil {
		return "", err
	}
	if fs, ok := b.fs.(afero.LinkReader); ok {
		return fs.ReadlinkIfPossible(name)
	}
	return "", billy.ErrNotSupported
}

func (b *Billy) Chmod(name string, mode os.FileMode) error {
	if err := local("chmod", name); err != nil {
		return err
	}
	return b.fs.Chmod(name, mode)
}

type billyFile struct {
	afero.File
	name string
}

func (f *billyFile) Name() string  { return f.name }
func (f *billyFile) Lock() error   { return lock(f.File, true) }
func (f *billyFile) Unlock() error { return lock(f.File, false) }
