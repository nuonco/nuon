package rootfs

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

// PrepareFS selects the pinned OS backend for nil or OsFs, and a private job
// namespace for MemMapFs. Other injected backends fail closed rather than
// accidentally treating a host filesystem as a relative namespace.
// On success the caller owns the OS parent and child handles, if present.
func PrepareFS(backing afero.Fs, parentPath, name string, replace bool) (*FS, afero.Fs, error) {
	switch backing.(type) {
	case nil, *afero.OsFs:
		parent, child, err := Prepare(parentPath, name, replace)
		if err != nil {
			return nil, nil, err
		}
		return parent, child, nil
	case *afero.MemMapFs:
		if name == "." || filepath.Base(name) != name || local("prepare", name) != nil {
			return nil, nil, fmt.Errorf("invalid directory name %q", name)
		}
		path := filepath.Join(parentPath, name)
		if replace {
			if err := backing.RemoveAll(path); err != nil {
				return nil, nil, err
			}
		}
		if err := backing.MkdirAll(path, 0o700); err != nil {
			return nil, nil, err
		}
		return nil, afero.NewBasePathFs(backing, path), nil
	default:
		return nil, nil, fmt.Errorf("unsupported workspace filesystem %T", backing)
	}
}

// RemoveFS removes one directory without following its final symlink, including
// after failed initialization has closed the job handles. parentPath is trusted.
func RemoveFS(backing afero.Fs, parentPath, name string) error {
	if name == "." || filepath.Base(name) != name || local("remove", name) != nil {
		return fmt.Errorf("invalid directory name %q", name)
	}
	switch backing.(type) {
	case nil, *afero.OsFs:
		canonical, err := filepath.EvalSymlinks(parentPath)
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return err
		}
		parent, err := Open(canonical)
		if err != nil {
			return err
		}
		defer parent.Close()
		return parent.RemoveAll(name)
	case *afero.MemMapFs:
		return backing.RemoveAll(filepath.Join(parentPath, name))
	default:
		return fmt.Errorf("unsupported workspace filesystem %T", backing)
	}
}

// Prepare creates a private single-component directory under a trusted parent.
// The caller owns both returned handles and must close both, including on later
// initialization failure. Existing symlink roots are rejected even with replace.
func Prepare(parentPath, name string, replace bool) (parent, child *FS, err error) {
	if name == "." || filepath.Base(name) != name || local("prepare", name) != nil {
		return nil, nil, fmt.Errorf("invalid directory name %q", name)
	}
	if err := os.MkdirAll(parentPath, 0o700); err != nil {
		return nil, nil, err
	}
	// The configured parent is trusted (and /tmp is an alias on macOS).
	parentPath, err = filepath.EvalSymlinks(parentPath)
	if err != nil {
		return nil, nil, err
	}
	parent, err = Open(parentPath)
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		if err != nil {
			parent.Close()
			if child != nil {
				child.Close()
			}
		}
	}()
	info, _, statErr := parent.LstatIfPossible(name)
	if statErr != nil && !os.IsNotExist(statErr) {
		return parent, nil, statErr
	}
	if statErr == nil {
		if !info.IsDir() {
			return parent, nil, fmt.Errorf("workspace root %q is not a directory", name)
		}
		if replace {
			if err := parent.RemoveAll(name); err != nil {
				return parent, nil, err
			}
		}
	}
	if err := parent.Mkdir(name, 0o700); err != nil && !os.IsExist(err) {
		return parent, nil, err
	}
	info, _, err = parent.LstatIfPossible(name)
	if err != nil || !info.IsDir() {
		return parent, nil, fmt.Errorf("workspace root %q changed while preparing", name)
	}
	child, err = parent.Sub(name)
	if err != nil {
		return parent, nil, err
	}
	pinned, err := child.Stat(".")
	if err != nil || !os.SameFile(info, pinned) {
		return parent, child, fmt.Errorf("workspace root %q changed while opening", name)
	}
	if err := child.Chmod(".", 0o700); err != nil {
		return parent, child, err
	}
	return parent, child, nil
}
