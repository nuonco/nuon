package builder

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

func (b *Builder) getSourceFiles(ctx context.Context, root string) ([]string, error) {
	fps := make([]string, 0)

	if err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		fps = append(fps, path)
		return nil
	}); err != nil {
		return nil, fmt.Errorf("unable to walk %s: %w", root, err)
	}

	return fps, nil
}
