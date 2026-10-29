package ociarchive

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nuonco/nuon/pkg/sourcefs"
)

// SourceFiles selects a directory within the repository and maps its inputs to
// directory-relative archive paths. Links outside that directory are materialized
// within the archive, but may never leave the supplied repository filesystem.
func SourceFiles(ctx context.Context, source *sourcefs.FS, root, fileType string) ([]FileRef, error) {
	if source == nil {
		return nil, fmt.Errorf("archive source filesystem is required")
	}
	root, err := source.Resolve(root)
	if err != nil {
		return nil, err
	}
	files := make([]FileRef, 0)
	ancestors := make(map[string]bool)
	var walk func(string, string) error
	walk = func(dir, relative string) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		resolved, err := source.Resolve(dir)
		if err != nil {
			return err
		}
		if ancestors[resolved] {
			return fmt.Errorf("source directory symlink cycle: %s", dir)
		}
		ancestors[resolved] = true
		defer delete(ancestors, resolved)
		entries, err := source.ReadDir(dir)
		if err != nil {
			return err
		}
		relocated := resolved != filepath.Join(root, relative)
		for _, info := range entries {
			name := filepath.Join(dir, info.Name())
			relativeName := filepath.Join(relative, info.Name())
			if info.IsDir() {
				if info.Name() == ".git" {
					continue
				}
				if err := walk(name, relativeName); err != nil {
					return err
				}
				continue
			}
			if info.Mode()&os.ModeSymlink != 0 {
				resolved, err := source.Resolve(name)
				if err != nil {
					return err
				}
				target, err := source.Lstat(resolved)
				if err != nil {
					return err
				}
				if target.IsDir() {
					link, err := source.Readlink(name)
					if err != nil {
						return err
					}
					targetRelative, err := filepath.Rel(root, resolved)
					if err != nil {
						return err
					}
					archiveTarget := filepath.Join(filepath.Dir(relativeName), link)
					if relocated || filepath.IsAbs(link) || !filepath.IsLocal(archiveTarget) || !filepath.IsLocal(targetRelative) {
						// Relocation can change a relative link's meaning even when
						// both targets are local; include its contents at the link.
						if err := walk(name, relativeName); err != nil {
							return err
						}
						continue
					}
				}
			}
			files = append(files, FileRef{
				AbsPath:  filepath.Join(source.Root(), name),
				RelPath:  filepath.ToSlash(relativeName),
				FileType: fileType,
			})
		}
		return nil
	}
	if err := walk(root, "."); err != nil {
		return nil, fmt.Errorf("unable to walk %s: %w", filepath.Join(source.Root(), root), err)
	}
	return files, nil
}
