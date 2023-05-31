package oci

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/powertoolsdev/mono/pkg/terraform/archive"
)

// Unpack fetches an archive, and calls the callback with each file contained within it
func (o *oci) Unpack(ctx context.Context, cb archive.Callback) error {
	if err := o.pull(ctx); err != nil {
		return fmt.Errorf("unable to pull archive: %w", err)
	}

	if err := o.unpackDir(ctx, cb); err != nil {
		return fmt.Errorf("unable to unpack directory: %w", err)
	}

	return nil
}

func (o *oci) unpackDir(ctx context.Context, cb archive.Callback) error {
	fn := func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		rc, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("unable to open file: %w", err)
		}

		relPath := strings.TrimPrefix(path, o.tmpDir+"/")
		if err := cb(ctx, relPath, rc); err != nil {
			return fmt.Errorf("unable to execute callback: %w", err)
		}

		return nil
	}

	if err := filepath.Walk(o.tmpDir, fn); err != nil {
		return fmt.Errorf("unable to walk root directory: %w", err)
	}

	return nil
}
