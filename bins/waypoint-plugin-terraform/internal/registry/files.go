package registry

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultFileType     string = "file/terraform"
	defaultArtifactType string = "artifact/terraform"
	defaultTag          string = "latest"
)

type fileRef struct {
	absPath string
	relPath string
}

func (b *Registry) getSourceFiles(ctx context.Context, root string) ([]fileRef, error) {
	fps := make([]fileRef, 0)

	if !strings.HasSuffix(root, "/") {
		root = root + "/"
	}
	if err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		fps = append(fps, fileRef{
			absPath: path,
			relPath: strings.TrimPrefix(path, root),
		})
		return nil
	}); err != nil {
		return nil, fmt.Errorf("unable to walk %s: %w", root, err)
	}

	return fps, nil
}
