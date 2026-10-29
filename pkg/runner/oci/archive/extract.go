package ociarchive

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/nuonco/nuon/pkg/rootfs"
	"github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/afero"
	"oras.land/oras-go/v2/content/file"
)

type directoryMode struct {
	mode     os.FileMode
	explicit bool
	cleanup  afero.File
}

type extraction struct {
	fs          *rootfs.FS
	directories map[string]directoryMode
	symlinks    []string
}

func (e *extraction) layer(ctx context.Context, desc v1.Descriptor, blob io.Reader) error {
	title := desc.Annotations[v1.AnnotationTitle]
	unpack := desc.Annotations[file.AnnotationUnpack] == "true"
	if err := validateTarPath(title); err != nil && !(unpack && title == ".") {
		return err
	}
	if !unpack {
		return e.regular(ctx, title, blob, 0644)
	}
	gz, err := gzip.NewReader(blob)
	if err != nil {
		return err
	}
	defer gz.Close()
	var reader io.Reader = gz
	var verifier digest.Verifier
	if checksum := desc.Annotations[file.AnnotationDigest]; checksum != "" {
		d, err := digest.Parse(checksum)
		if err != nil {
			return fmt.Errorf("invalid uncompressed digest: %w", err)
		}
		verifier = d.Verifier()
		reader = io.TeeReader(reader, verifier)
	}
	tr := tar.NewReader(reader)
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		name := header.Name
		if header.Typeflag == tar.TypeDir {
			name = strings.TrimSuffix(name, "/")
		}
		if err := validateTarPath(name); err != nil && !(name == "." && header.Typeflag == tar.TypeDir) {
			return err
		}
		if title != "." && name != title && !strings.HasPrefix(name, title+"/") {
			return fmt.Errorf("archive entry %q is outside layer %q", name, title)
		}
		switch header.Typeflag {
		case tar.TypeReg, tar.TypeRegA:
			err = e.regular(ctx, name, tr, header.FileInfo().Mode().Perm())
		case tar.TypeDir:
			err = e.directory(name, header.FileInfo().Mode().Perm(), true)
		case tar.TypeSymlink:
			err = e.symlink(name, header.Linkname)
		case tar.TypeLink:
			return fmt.Errorf("archive hardlinks are not supported: %q", name)
		default:
			return fmt.Errorf("unsupported archive entry type %d: %q", header.Typeflag, name)
		}
		if err != nil {
			return err
		}
	}
	// tar stops at its end records, before gzip's trailer (and often padding).
	// Drain to validate the gzip checksum and the complete uncompressed digest.
	if _, err := io.Copy(io.Discard, &contextReader{ctx: ctx, reader: reader}); err != nil {
		return err
	}
	if verifier != nil && !verifier.Verified() {
		return fmt.Errorf("uncompressed content digest mismatch")
	}
	return nil
}

func (e *extraction) regular(ctx context.Context, name string, reader io.Reader, mode os.FileMode) error {
	if err := e.directory(path.Dir(name), 0755, false); err != nil {
		return err
	}
	// Exclusive creation neither follows terminal symlinks nor truncates a
	// preexisting hardlink. os.Root confines the entire lookup during races.
	out, err := e.fs.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return fmt.Errorf("unable to create archive file %q: %w", name, err)
	}
	_, copyErr := io.Copy(out, &contextReader{ctx: ctx, reader: reader})
	modeErr := setFileMode(out, mode)
	closeErr := out.Close()
	if copyErr != nil {
		return copyErr
	}
	if modeErr != nil {
		return modeErr
	}
	return closeErr
}

func (e *extraction) directory(name string, mode os.FileMode, explicit bool) error {
	if name != "." {
		if err := e.directory(path.Dir(name), 0755, false); err != nil {
			return err
		}
		info, _, err := e.fs.LstatIfPossible(name)
		if os.IsNotExist(err) {
			if err := e.fs.Mkdir(name, 0700); err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else if !info.IsDir() {
			return fmt.Errorf("archive directory %q is not a directory (symlinks are not allowed)", name)
		}
	}
	previous, ok := e.directories[name]
	if ok && previous.explicit {
		if explicit && previous.mode != mode {
			return fmt.Errorf("conflicting archive directory modes for %q", name)
		}
		return nil
	}
	e.directories[name] = directoryMode{mode: mode, explicit: explicit}
	return nil
}

func (e *extraction) symlink(name, target string) error {
	if err := validateSymlinkTarget(name, target); err != nil {
		return err
	}
	if err := e.directory(path.Dir(name), 0755, false); err != nil {
		return err
	}
	if err := e.fs.SymlinkIfPossible(target, name); err != nil {
		return err
	}
	e.symlinks = append(e.symlinks, name)
	return nil
}

func validateSymlinkTarget(name, target string) error {
	if target == "" || path.IsAbs(target) || strings.ContainsAny(target, "\\\x00") || len(target) >= 2 && target[1] == ':' {
		return fmt.Errorf("unsafe archive symlink %q -> %q", name, target)
	}
	return nil
}

func (e *extraction) validateSymlinks(ctx context.Context) error {
	// Forward references can span layers. Validate only after all entries exist;
	// directory() and exclusive file creation forbid writes through these links.
	for _, name := range e.symlinks {
		pending := strings.Split(name, "/")
		resolved := "."
		links := 0
		for len(pending) > 0 {
			if err := ctx.Err(); err != nil {
				return err
			}
			part := pending[0]
			pending = pending[1:]
			switch part {
			case "", ".":
				continue
			case "..":
				if resolved == "." {
					return fmt.Errorf("archive symlink escapes root: %q", name)
				}
				resolved = path.Dir(resolved)
				continue
			}
			next := path.Join(resolved, part)
			if err := validateTarPath(next); err != nil {
				return err
			}
			info, _, err := e.fs.LstatIfPossible(next)
			if os.IsNotExist(err) {
				// Legacy layers may omit empty directories, leaving confined links
				// dangling. Their remaining components still need escape checks.
				resolved = next
				continue
			}
			if err != nil {
				return err
			}
			if info.Mode()&os.ModeSymlink != 0 {
				links++
				if links > 255 {
					return fmt.Errorf("archive symlink loop or excessive link depth: %q", name)
				}
				target, err := e.fs.ReadlinkIfPossible(next)
				if err != nil {
					return err
				}
				if err := validateSymlinkTarget(next, target); err != nil {
					return err
				}
				// Expand links before cleaning "..": a/.. escapes if a points to ".".
				pending = append(strings.Split(target, "/"), pending...)
				continue
			}
			if len(pending) > 0 && !info.IsDir() {
				return fmt.Errorf("archive symlink component is not a directory: %q", next)
			}
			resolved = next
		}
	}
	return nil
}

func (e *extraction) restoreDirectoryModes() error {
	names := make([]string, 0, len(e.directories))
	for name := range e.directories {
		names = append(names, name)
	}
	// Children must be finished before removing search/write access on parents.
	sort.Sort(sort.Reverse(sort.StringSlice(names)))
	for _, name := range names {
		dir, err := e.fs.Open(name)
		if err != nil {
			return err
		}
		info, err := dir.Stat()
		if err == nil && !info.IsDir() {
			err = fmt.Errorf("archive directory %q changed during extraction", name)
		}
		if err == nil {
			err = setFileMode(dir, e.directories[name].mode)
		}
		if err != nil {
			_ = dir.Close()
			return err
		}
		metadata := e.directories[name]
		if metadata.mode&0500 != 0500 {
			// Without owner read/search permission this directory cannot be
			// reopened for cleanup, even through a pinned os.Root.
			metadata.cleanup = dir
			e.directories[name] = metadata
		} else if err := dir.Close(); err != nil {
			return err
		}
	}
	return nil
}

func setFileMode(file afero.File, mode os.FileMode) error {
	// Never re-resolve a pathname for chmod: it could now name a hardlink.
	chmod, ok := file.(interface{ Chmod(os.FileMode) error })
	if !ok {
		return fmt.Errorf("archive file does not support handle-based chmod")
	}
	return chmod.Chmod(mode)
}

type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r *contextReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.reader.Read(p)
}
