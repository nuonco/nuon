package sourcefs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"

	"github.com/nuonco/nuon/pkg/rootfs"
)

type FS struct {
	fs    afero.Fs
	root  string
	base  string
	close func() error
}

type Link struct {
	Path, Target string
	Absolute     bool
}

// New pins an OS source root or scopes an in-memory filesystem. Other backends
// must already be confined and explicitly borrowed with From.
func New(fs afero.Fs, root string) (*FS, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	base := root
	if parent, ok := fs.(*rootfs.FS); ok {
		source, err := From(parent, parent.Root())
		if err != nil {
			return nil, err
		}
		return source.Sub(root)
	}
	switch fs.(type) {
	case *afero.OsFs, *afero.MemMapFs:
	default:
		return nil, fmt.Errorf("unsupported source filesystem %T", fs)
	}
	root, _, err = resolve(fs, filepath.VolumeName(root)+string(filepath.Separator), root, "")
	if err != nil {
		return nil, err
	}
	info, err := fs.Stat(root)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("source root is not a directory: %s", root)
	}
	if _, ok := fs.(*afero.OsFs); ok {
		pinned, err := rootfs.Open(root)
		if err != nil {
			return nil, err
		}
		return &FS{fs: pinned, root: root, base: base, close: pinned.Close}, nil
	}
	return &FS{fs: afero.NewBasePathFs(fs, root), root: root, base: base}, nil
}

// From borrows an already-rooted filesystem. The root label is used only to
// interpret absolute input names; all filesystem operations remain relative.
func From(fs afero.Fs, root string) (*FS, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	info, err := fs.Stat(".")
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("source root is not a directory: %s", root)
	}
	return &FS{fs: fs, root: root, base: root}, nil
}

func (f *FS) Close() error {
	if f != nil && f.close != nil {
		return f.close()
	}
	return nil
}

func (f *FS) Sub(name string) (*FS, error) {
	rel, _, err := f.Resolve(name)
	if err != nil {
		return nil, err
	}
	if parent, ok := f.fs.(*rootfs.FS); ok {
		pinned, err := parent.Sub(rel)
		if err != nil {
			return nil, err
		}
		return &FS{fs: pinned, root: pinned.Root(), base: filepath.Join(f.base, rel), close: pinned.Close}, nil
	}
	if rel == "." {
		return From(f.fs, f.root)
	}
	return From(afero.NewBasePathFs(f.fs, rel), filepath.Join(f.root, rel))
}

func (f *FS) Root() string { return f.root }

func (f *FS) Relative(name string) (string, error) {
	if filepath.IsAbs(name) {
		if within(f.base, name) {
			return filepath.Rel(f.base, name)
		}
	} else {
		name = filepath.Clean(name)
		if name == ".." || strings.HasPrefix(name, ".."+string(filepath.Separator)) {
			return "", fmt.Errorf("source path escapes repository: %s", name)
		}
		return name, nil
	}
	if !within(f.root, name) {
		return "", fmt.Errorf("source path escapes repository: %s", name)
	}
	return filepath.Rel(f.root, name)
}

func (f *FS) Resolve(name string) (string, []Link, error) {
	return f.ResolveWithin(".", name)
}

func (f *FS) ResolveWithin(root, name string) (string, []Link, error) {
	rel, err := f.Relative(name)
	if err != nil {
		return "", nil, err
	}
	scope, err := f.Relative(root)
	if err != nil {
		return "", nil, err
	}
	resolvedScope, _, err := resolve(f.fs, ".", scope, f.root, f.base)
	if err != nil {
		return "", nil, err
	}
	if suffix, err := filepath.Rel(scope, rel); err == nil && filepath.IsLocal(suffix) {
		rel = filepath.Join(resolvedScope, suffix)
	}
	return resolve(f.fs, resolvedScope, rel, f.root, f.base)
}

func (f *FS) Lstat(name string) (os.FileInfo, error) {
	name, err := f.linkPath(name)
	if err != nil {
		return nil, err
	}
	return lstat(f.fs, name)
}

func (f *FS) Readlink(name string) (string, error) {
	name, err := f.linkPath(name)
	if err != nil {
		return "", err
	}
	return readlink(f.fs, name)
}

func (f *FS) linkPath(name string) (string, error) {
	rel, err := f.Relative(name)
	if err != nil {
		return "", err
	}
	if rel == "." {
		return ".", nil
	}
	parent, _, err := f.Resolve(filepath.Dir(rel))
	if err != nil {
		return "", err
	}
	return filepath.Join(parent, filepath.Base(rel)), nil
}

func (f *FS) Open(name string) (afero.File, error) {
	resolved, _, err := f.Resolve(name)
	if err != nil {
		return nil, err
	}
	return f.fs.Open(resolved)
}

func (f *FS) ReadFile(name string) ([]byte, error) {
	resolved, _, err := f.Resolve(name)
	if err != nil {
		return nil, err
	}
	return afero.ReadFile(f.fs, resolved)
}

func (f *FS) ReadDir(name string) ([]os.FileInfo, error) {
	resolved, _, err := f.Resolve(name)
	if err != nil {
		return nil, err
	}
	return afero.ReadDir(f.fs, resolved)
}

func (f *FS) Glob(pattern string) ([]string, error) {
	rel, err := f.Relative(pattern)
	if err != nil {
		return nil, err
	}
	return afero.Glob(f.fs, rel)
}

func (f *FS) Walk(name string, fn filepath.WalkFunc) error {
	rel, err := f.Relative(name)
	if err != nil {
		return err
	}
	return afero.Walk(f.fs, rel, fn)
}

func within(root, name string) bool {
	rel, err := filepath.Rel(root, name)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}

func lstat(fs afero.Fs, name string) (os.FileInfo, error) {
	if fs, ok := fs.(afero.Lstater); ok {
		info, _, err := fs.LstatIfPossible(name)
		return info, err
	}
	return nil, fmt.Errorf("source filesystem does not support lstat")
}

func readlink(fs afero.Fs, name string) (string, error) {
	if fs, ok := fs.(afero.LinkReader); ok {
		return fs.ReadlinkIfPossible(name)
	}
	return "", fmt.Errorf("source filesystem does not support readlink")
}

func resolve(fs afero.Fs, root, name string, absoluteRoots ...string) (string, []Link, error) {
	if !within(root, name) {
		return "", nil, fmt.Errorf("source path escapes repository: %s", name)
	}
	rel, err := filepath.Rel(root, name)
	if err != nil {
		return "", nil, err
	}
	pending := strings.Split(rel, string(filepath.Separator))
	resolved := root
	var links []Link
	for len(pending) > 0 {
		part := pending[0]
		pending = pending[1:]
		if part == "" || part == "." {
			continue
		}
		if part == ".." {
			if resolved == root {
				return "", nil, fmt.Errorf("source symlink escapes repository: %s", name)
			}
			resolved = filepath.Dir(resolved)
			continue
		}
		next := filepath.Join(resolved, part)
		info, err := lstat(fs, next)
		if err != nil {
			return "", nil, err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			if len(links) >= 255 {
				return "", nil, fmt.Errorf("source symlink loop or excessive link depth: %s", name)
			}
			target, err := readlink(fs, next)
			if err != nil {
				return "", nil, err
			}
			linkPath, err := filepath.Rel(root, next)
			if err != nil {
				return "", nil, err
			}
			absolute := filepath.IsAbs(target)
			if absolute {
				matched := false
				for _, absoluteRoot := range absoluteRoots {
					boundary := root
					if !filepath.IsAbs(boundary) {
						boundary = filepath.Join(absoluteRoot, root)
					}
					prefix := strings.TrimRight(boundary, string(filepath.Separator)) + string(filepath.Separator)
					if target == boundary || strings.HasPrefix(target, prefix) {
						target = strings.TrimPrefix(target, boundary)
						target = strings.TrimPrefix(target, string(filepath.Separator))
						matched = true
						break
					}
				}
				if !matched {
					return "", nil, fmt.Errorf("source symlink escapes repository: %s", next)
				}
				resolved = root
			}
			links = append(links, Link{Path: filepath.ToSlash(linkPath), Target: filepath.ToSlash(target), Absolute: absolute})
			pending = append(strings.Split(target, string(filepath.Separator)), pending...)
			continue
		}
		if len(pending) > 0 && !info.IsDir() {
			return "", nil, fmt.Errorf("source path component is not a directory: %s", next)
		}
		resolved = next
	}
	return resolved, links, nil
}
