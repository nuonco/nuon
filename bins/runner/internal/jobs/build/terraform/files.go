package terraform

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	ociarchive "github.com/powertoolsdev/mono/bins/runner/internal/pkg/oci/archive"
)

func (h *handler) getSourceFiles(ctx context.Context, root string) ([]ociarchive.FileRef, error) {
	fps := make([]ociarchive.FileRef, 0)

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

		fps = append(fps, ociarchive.FileRef{
			AbsPath:  path,
			RelPath:  strings.TrimPrefix(path, root),
			FileType: defaultFileType,
		})
		return nil
	}); err != nil {
		return nil, fmt.Errorf("unable to walk %s: %w", root, err)
	}

	return fps, nil
}
