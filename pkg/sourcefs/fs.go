package sourcefs

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/spf13/afero"

	"github.com/nuonco/nuon/pkg/rootfs"
)

var errGitMetadata = errors.New("Git metadata is not authored source")

type FS struct {
	fs     afero.Fs
	root   string
	base   string
	close  func() error
	reads  *reads
	prefix string
}

type reads struct {
	mu    sync.Mutex
	paths map[string]bool
}

// Reads lists repository paths whose contents were read or listed, and paths
// that were only resolved, statted, or found missing.
type Reads struct {
	Read, Stat []string
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
	rel, err := f.resolve(name)
	if err != nil {
		return nil, err
	}
	if parent, ok := f.fs.(*rootfs.FS); ok {
		pinned, err := parent.Sub(rel)
		if err != nil {
			return nil, err
		}
		return &FS{fs: pinned, root: pinned.Root(), base: filepath.Join(f.base, rel), close: pinned.Close, reads: f.reads, prefix: filepath.Join(f.prefix, rel)}, nil
	}
	var sub *FS
	if rel == "." {
		sub, err = From(f.fs, f.root)
	} else {
		sub, err = From(afero.NewBasePathFs(f.fs, rel), filepath.Join(f.root, rel))
	}
	if err != nil {
		return nil, err
	}
	sub.reads, sub.prefix = f.reads, filepath.Join(f.prefix, rel)
	return sub, nil
}

// Record starts collecting the paths this filesystem and its views touch.
func (f *FS) Record() {
	f.reads = &reads{paths: make(map[string]bool)}
}

func (f *FS) Reads() Reads {
	var out Reads
	if f.reads == nil {
		return out
	}
	f.reads.mu.Lock()
	defer f.reads.mu.Unlock()
	for name, read := range f.reads.paths {
		if read {
			out.Read = append(out.Read, name)
		} else {
			out.Stat = append(out.Stat, name)
		}
	}
	sort.Strings(out.Read)
	sort.Strings(out.Stat)
	return out
}

func (f *FS) record(name string, read bool) {
	if f.reads == nil {
		return
	}
	name = filepath.ToSlash(filepath.Join(f.prefix, name))
	f.reads.mu.Lock()
	f.reads.paths[name] = f.reads.paths[name] || read
	f.reads.mu.Unlock()
}

func (f *FS) stat(name string) { f.record(name, false) }

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
	resolved, err := f.resolve(name)
	if err == nil {
		f.stat(resolved)
	}
	return resolved, err
}

func (f *FS) resolve(name string) (string, error) {
	rel, err := f.Relative(name)
	if err != nil {
		return "", err
	}
	return resolveTrace(f.fs, ".", rel, nil, f.visitor(), f.root, f.base)
}

func (f *FS) visitor() func(string) {
	if f.reads == nil {
		return nil
	}
	return f.stat
}

func (f *FS) ResolveWithLinks(name string) (string, []Link, error) {
	rel, err := f.Relative(name)
	if err != nil {
		return "", nil, err
	}
	var links []Link
	resolved, err := resolveTrace(f.fs, ".", rel, &links, f.visitor(), f.root, f.base)
	if err == nil {
		f.stat(resolved)
	}
	return resolved, links, err
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

// Entry is a directory entry as stored. Target is a symlink's text.
type Entry struct {
	os.FileInfo
	Target string
}

// Entry is Lstat plus Readlink of name. Unlike Lstat, it accepts a symlink
// whose target resolves into Git metadata: Git tracks such a link as its text,
// and reading that text never follows the link.
func (f *FS) Entry(name string) (Entry, error) {
	rel, err := f.Relative(name)
	if err != nil {
		return Entry{}, err
	}
	if rel != "." {
		parent, err := f.resolve(filepath.Dir(rel))
		if err != nil {
			return Entry{}, err
		}
		rel = filepath.Join(parent, filepath.Base(rel))
	}
	f.stat(rel)
	return entry(f.fs, rel)
}

// Entries lists a directory like ReadDir, with the symlink exception of Entry.
func (f *FS) Entries(name string) ([]Entry, error) {
	resolved, err := f.resolve(name)
	if err != nil {
		return nil, err
	}
	f.record(resolved, true)
	infos, err := afero.ReadDir(f.fs, resolved)
	if err != nil {
		return nil, err
	}
	var out []Entry
	for _, info := range infos {
		child := filepath.Join(resolved, info.Name())
		if _, err := f.Relative(child); err != nil {
			continue
		}
		f.stat(child)
		e, err := entry(f.fs, child)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, nil
}

func entry(fs afero.Fs, name string) (Entry, error) {
	info, err := lstat(fs, name)
	if err != nil {
		return Entry{}, err
	}
	out := Entry{FileInfo: info}
	if info.Mode()&os.ModeSymlink != 0 {
		out.Target, err = readlink(fs, name)
	}
	return out, err
}

func (f *FS) linkPath(name string) (string, error) {
	rel, err := f.Relative(name)
	if err != nil {
		return "", err
	}
	if rel == "." {
		f.stat(".")
		return ".", nil
	}
	if _, err := f.resolve(rel); errors.Is(err, errGitMetadata) {
		return "", err
	}
	parent, err := f.resolve(filepath.Dir(rel))
	if err != nil {
		return "", err
	}
	name = filepath.Join(parent, filepath.Base(rel))
	f.stat(name)
	return name, nil
}

func (f *FS) Open(name string) (afero.File, error) {
	resolved, err := f.resolve(name)
	if err != nil {
		return nil, err
	}
	f.record(resolved, true)
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
			if _, err := d.source.resolve(filepath.Join(d.name, entry.Name())); !errors.Is(err, errGitMetadata) {
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
	resolved, err := f.resolve(name)
	if err != nil {
		return nil, err
	}
	f.record(resolved, true)
	return afero.ReadFile(f.fs, resolved)
}

func (f *FS) ReadDir(name string) ([]os.FileInfo, error) {
	resolved, err := f.resolve(name)
	if err != nil {
		return nil, err
	}
	f.record(resolved, true)
	entries, err := afero.ReadDir(f.fs, resolved)
	if err != nil {
		return nil, err
	}
	visible := entries[:0]
	for _, entry := range entries {
		if _, err := f.resolve(filepath.Join(resolved, entry.Name())); !errors.Is(err, errGitMetadata) {
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
	matches, err := f.glob(rel)
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

// glob matches like afero.Glob and records each directory it lists, so a
// wildcard that crosses a link records where the link leads even when nothing
// matches there.
func (f *FS) glob(pattern string) ([]string, error) {
	if !hasMeta(pattern) {
		if resolved, err := f.resolve(pattern); err == nil {
			f.stat(resolved)
		}
		if _, err := lstat(f.fs, pattern); err != nil {
			return nil, nil
		}
		return []string{pattern}, nil
	}
	dir, file := filepath.Split(pattern)
	dir = filepath.Clean(dir)
	dirs := []string{dir}
	if hasMeta(dir) {
		var err error
		if dirs, err = f.glob(dir); err != nil {
			return nil, err
		}
	}
	var matches []string
	for _, dir := range dirs {
		resolved, err := f.resolve(dir)
		if err != nil {
			continue
		}
		info, err := lstat(f.fs, resolved)
		if err != nil || !info.IsDir() {
			f.stat(resolved)
			continue
		}
		f.record(resolved, true)
		listing, err := f.fs.Open(resolved)
		if err != nil {
			continue
		}
		names, _ := listing.Readdirnames(-1)
		if err := listing.Close(); err != nil {
			return nil, err
		}
		sort.Strings(names)
		for _, name := range names {
			matched, err := filepath.Match(file, name)
			if err != nil {
				return nil, err
			}
			if matched {
				matches = append(matches, filepath.Join(dir, name))
			}
		}
	}
	return matches, nil
}

func (f *FS) Walk(name string, fn filepath.WalkFunc) error {
	rel, err := f.Relative(name)
	if err != nil {
		return err
	}
	if resolved, err := f.resolve(rel); err == nil {
		f.record(resolved, true)
	}
	return afero.Walk(f.fs, rel, func(name string, info os.FileInfo, err error) error {
		if err == nil {
			if _, resolveErr := f.resolve(name); errors.Is(resolveErr, errGitMetadata) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}
		return fn(name, info, err)
	})
}

func hasMeta(pattern string) bool { return strings.ContainsAny(pattern, "*?[") }

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
	return resolveTrace(fs, root, name, nil, nil, absoluteRoots...)
}

// visit receives each symlink followed and the component that ended a failed
// lookup, so a recorder sees every entry that shaped the result.
func resolveTrace(fs afero.Fs, root, name string, links *[]Link, visit func(string), absoluteRoots ...string) (string, error) {
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
		if visit != nil && (err != nil || info.Mode()&os.ModeSymlink != 0 || len(pending) > 0 && !info.IsDir()) {
			visit(next)
		}
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
					return "", fmt.Errorf("source symlink escapes repository: %s", next)
				}
				resolved = root
			}
			if links != nil {
				linkPath, err := filepath.Rel(root, next)
				if err != nil {
					return "", err
				}
				*links = append(*links, Link{Path: filepath.ToSlash(linkPath), Target: filepath.ToSlash(target), Absolute: absolute})
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
