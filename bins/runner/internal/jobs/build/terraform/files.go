package terraform

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

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

const (
	backendBlock string = "backend s3 {}"
)

func (h *handler) validateSourceFiles(ctx context.Context, files []ociarchive.FileRef) error {
	for _, src := range files {
		byts, err := os.ReadFile(src.AbsPath)
		if err != nil {
			return errors.Wrap(err, "unable to read file")
		}

		if bytes.Contains(byts, []byte(backendBlock)) {
			return nil
		}
	}

	return errors.New("no backend s3 {} block found")
}
