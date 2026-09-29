package dir

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

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

func (p *parser) listDirFrom(path string, ancestors []os.FileInfo) ([]string, error) {
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

		info := os.FileInfo(entry)
		if entry.Mode()&os.ModeSymlink != 0 {
			target, err := p.fs.Stat(fp)
			if err != nil {
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

func isAncestor(info os.FileInfo, ancestors []os.FileInfo) bool {
	for _, ancestor := range ancestors {
		if os.SameFile(ancestor, info) {
			return true
		}
	}

	return false
}
