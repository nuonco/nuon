package workspace

import (
	"context"
	"fmt"
	"io"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/terraform/hooks"
)

// LoadArchive loads the archives into the workspace
func (w *workspace) LoadArchive(ctx context.Context) error {
	if err := w.Archive.Init(ctx); err != nil {
		return fmt.Errorf("unable to initialize archive: %w", err)
	}

	// NOTE(jm): this isn't the most efficient way of writing each file, but since most of our files will just be
	// source code files it probably isn't hurting anything at the moment.
	cb := func(_ context.Context, name string, reader io.ReadCloser) error {
		byts, err := io.ReadAll(reader)
		if err != nil {
			return fmt.Errorf("unable to read file in callback: %w", err)
		}
		defer reader.Close()

		permissions := defaultFilePermissions
		if generics.SliceContains(name, hooks.ValidHooks()) {
			permissions = defaultFileExecPermissions
		}

		if err := w.writeFile(name, byts, permissions); err != nil {
			return fmt.Errorf("unable to write file: %w", err)
		}
		return nil
	}

	if err := w.Archive.Unpack(ctx, cb); err != nil {
		return fmt.Errorf("unable to unpack archive: %w", err)
	}

	return nil
}
