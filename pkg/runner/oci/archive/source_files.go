package ociarchive

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nuonco/nuon/pkg/sourcefs"
)

// SourceFiles enumerates source inputs without following directory symlinks.
// Directory links are retained for the tar writer; .git directories are excluded.
func SourceFiles(ctx context.Context, source *sourcefs.FS, fileType string) ([]FileRef, error) {
	if source == nil {
		return nil, fmt.Errorf("archive source filesystem is required")
	}
	files := make([]FileRef, 0)
	var walk func(string) error
	walk = func(dir string) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		entries, err := source.ReadDir(dir)
		if err != nil {
			return err
		}
		for _, info := range entries {
			name := filepath.Join(dir, info.Name())
			if info.IsDir() {
				if info.Name() == ".git" {
					continue
				}
				if err := walk(name); err != nil {
					return err
				}
				continue
			}
			if info.Mode()&os.ModeSymlink != 0 {
				if _, _, err := source.Resolve(name); err != nil {
					return err
				}
			}
			files = append(files, FileRef{
				AbsPath:  filepath.Join(source.Root(), name),
				RelPath:  filepath.ToSlash(name),
				FileType: fileType,
			})
		}
		return nil
	}
	if err := walk("."); err != nil {
		return nil, fmt.Errorf("unable to walk %s: %w", source.Root(), err)
	}
	return files, nil
}
