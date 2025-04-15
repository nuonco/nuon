package dir

import (
	"os"

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
		if entry.IsDir() {
			continue
		}

		if !p.hasExtension(entry.Name()) {
			continue
		}

		files = append(files, entry.Name())
	}

	return files, nil
}
