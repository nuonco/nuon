package sourcefs

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"

	"github.com/nuonco/nuon/pkg/rootfs"
)

var errGitMetadata = errors.New("Git metadata is not authored source")

type FS struct {
	fs    afero.Fs
	root  string
	base  string
	close func() error
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
	root, err = resolve(fs, filepath.VolumeName(root)+string(filepath.Separator), root, "")
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
	rel, err := f.Resolve(name)
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
		root := f.root
		if within(f.base, name) {
			root = f.base
		}
		if !within(root, name) {
			return "", fmt.Errorf("source path escapes repository: %s", name)
		}
		var err error
		name, err = filepath.Rel(root, name)
		if err != nil {
			return "", err
		}
	} else {
		name = filepath.Clean(name)
		if name == ".." || strings.HasPrefix(name, ".."+string(filepath.Separator)) {
			return "", fmt.Errorf("source path escapes repository: %s", name)
		}
	}
	for _, part := range strings.Split(name, string(filepath.Separator)) {
		if strings.EqualFold(part, ".git") {
			return "", fmt.Errorf("%w: %s", errGitMetadata, name)
		}
	}
	return name, nil
}

func (f *FS) Resolve(name string) (string, error) {
	rel, err := f.Relative(name)
	if err != nil {
		return "", err
	}
	return resolve(f.fs, ".", rel, f.root, f.base)
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
	if _, err := f.Resolve(rel); errors.Is(err, errGitMetadata) {
		return "", err
	}
	parent, err := f.Resolve(filepath.Dir(rel))
	if err != nil {
		return "", err
	}
	return filepath.Join(parent, filepath.Base(rel)), nil
}

func (f *FS) Open(name string) (afero.File, error) {
	resolved, err := f.Resolve(name)
	if err != nil {
		return nil, err
	}
	file, err := f.fs.Open(resolved)
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil {
		return nil, errors.Join(err, file.Close())
	}
	if info.IsDir() {
		return &sourceDirectory{File: file, source: f, name: resolved}, nil
	}
	return file, nil
}

type sourceDirectory struct {
	afero.File
	source *FS
	name   string
}

func (d *sourceDirectory) Readdir(n int) ([]os.FileInfo, error) {
	var visible []os.FileInfo
	for {
		count := n
		if n > 0 {
			count -= len(visible)
		}
		entries, err := d.File.Readdir(count)
		for _, entry := range entries {
			if _, err := d.source.Resolve(filepath.Join(d.name, entry.Name())); !errors.Is(err, errGitMetadata) {
				visible = append(visible, entry)
			}
		}
		if err != nil || n <= 0 || len(visible) == n {
			return visible, err
		}
	}
}

func (d *sourceDirectory) Readdirnames(n int) ([]string, error) {
	entries, err := d.Readdir(n)
	names := make([]string, len(entries))
	for i, entry := range entries {
		names[i] = entry.Name()
	}
	return names, err
}

func (f *FS) ReadFile(name string) ([]byte, error) {
	resolved, err := f.Resolve(name)
	if err != nil {
		return nil, err
	}
	return afero.ReadFile(f.fs, resolved)
}

func (f *FS) ReadDir(name string) ([]os.FileInfo, error) {
	resolved, err := f.Resolve(name)
	if err != nil {
		return nil, err
	}
	entries, err := afero.ReadDir(f.fs, resolved)
	if err != nil {
		return nil, err
	}
	visible := entries[:0]
	for _, entry := range entries {
		if _, err := f.Resolve(filepath.Join(resolved, entry.Name())); !errors.Is(err, errGitMetadata) {
			visible = append(visible, entry)
		}
	}
	return visible, nil
}

func (f *FS) Glob(pattern string) ([]string, error) {
	rel, err := f.Relative(pattern)
	if err != nil {
		return nil, err
	}
	matches, err := afero.Glob(f.fs, rel)
	if err != nil {
		return nil, err
	}
	visible := matches[:0]
	for _, name := range matches {
		if _, err := f.Resolve(name); err == nil {
			visible = append(visible, name)
		}
	}
	return visible, nil
}

func (f *FS) Walk(name string, fn filepath.WalkFunc) error {
	rel, err := f.Relative(name)
	if err != nil {
		return err
	}
	return afero.Walk(f.fs, rel, func(name string, info os.FileInfo, err error) error {
		if err == nil {
			if _, resolveErr := f.Resolve(name); errors.Is(resolveErr, errGitMetadata) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}
		return fn(name, info, err)
	})
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

func resolve(fs afero.Fs, root, name string, absoluteRoots ...string) (string, error) {
	if !within(root, name) {
		return "", fmt.Errorf("source path escapes repository: %s", name)
	}
	rel, err := filepath.Rel(root, name)
	if err != nil {
		return "", err
	}
	pending := strings.Split(rel, string(filepath.Separator))
	resolved := root
	depth := 0
	for len(pending) > 0 {
		part := pending[0]
		pending = pending[1:]
		if part == "" || part == "." {
			continue
		}
		if part == ".." {
			if resolved == root {
				return "", fmt.Errorf("source symlink escapes repository: %s", name)
			}
			resolved = filepath.Dir(resolved)
			continue
		}
		if root == "." && strings.EqualFold(part, ".git") {
			return "", fmt.Errorf("%w: %s", errGitMetadata, name)
		}
		next := filepath.Join(resolved, part)
		info, err := lstat(fs, next)
		if err != nil {
			return "", err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			if depth >= 255 {
				return "", fmt.Errorf("source symlink loop or excessive link depth: %s", name)
			}
			depth++
			target, err := readlink(fs, next)
			if err != nil {
				return "", err
			}
			if filepath.IsAbs(target) {
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
					return "", fmt.Errorf("source symlink escapes repository: %s", next)
				}
				resolved = root
			}
			pending = append(strings.Split(target, string(filepath.Separator)), pending...)
			continue
		}
		if len(pending) > 0 && !info.IsDir() {
			return "", fmt.Errorf("source path component is not a directory: %s", next)
		}
		resolved = next
	}
	return resolved, nil
}
