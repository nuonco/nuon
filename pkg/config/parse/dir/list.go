package dir

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

func (p *parser) listDir(path string) ([]string, error) {
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

		if entry.IsDir() {
			subDirFiles, err := p.listDir(fp)
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
