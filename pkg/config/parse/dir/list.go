package dir

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

// maxDirDepth bounds how deep listDir will recurse. The ancestor check below
// catches symlink loops exactly on filesystems whose FileInfo supports
// os.SameFile; this is the backstop for the ones that do not.
const maxDirDepth = 64

func (p *parser) listDir(path string) ([]string, error) {
	info, err := p.fs.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, errors.Wrap(err, "unable to stat directory")
	}

	return p.listDirFrom(path, []os.FileInfo{info})
}

// listDirFrom collects every file under path carrying the configured
// extension. Symlinked directories are followed, so a config tree can share a
// directory with its siblings (components/images -> ../../shared/components/images).
//
// ancestors holds the FileInfo of each directory on the current recursion
// path, so a symlink pointing back up the tree is skipped rather than followed
// forever.
func (p *parser) listDirFrom(path string, ancestors []os.FileInfo) ([]string, error) {
	// Read directory entries
	entries, err := p.fs.ReadDir(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, errors.Wrap(err, "unable to read directory")
	}

	var files []string
	for _, entry := range entries {
		fp := filepath.Join(path, entry.Name())

		// ReadDir lstats its entries, so a symlink to a directory reports
		// itself rather than its target. Stat follows the link to find out
		// which it is.
		info := os.FileInfo(entry)
		if entry.Mode()&os.ModeSymlink != 0 {
			target, err := p.fs.Stat(fp)
			if err != nil {
				// Dangling or unreadable link; there is nothing to collect.
				continue
			}

			info = target
		}

		if info.IsDir() {
			if len(ancestors) >= maxDirDepth || isAncestor(info, ancestors) {
				continue
			}

			subDirFiles, err := p.listDirFrom(fp, append(ancestors, info))
			if err != nil {
				return nil, err
			}

			files = append(files, subDirFiles...)
			continue
		}

		if !p.hasExtension(entry.Name()) {
			continue
		}

		files = append(files, fp)
	}

	return files, nil
}

// isAncestor reports whether info is a directory already on the recursion
// path. It compares identity rather than paths, so a loop reached through
// differently-spelled symlinks is still caught.
func isAncestor(info os.FileInfo, ancestors []os.FileInfo) bool {
	for _, ancestor := range ancestors {
		if os.SameFile(ancestor, info) {
			return true
		}
	}

	return false
}
