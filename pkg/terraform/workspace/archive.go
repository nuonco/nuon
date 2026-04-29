package workspace

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/terraform/hooks"
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

		name = strings.TrimPrefix(name, "./")

		permissions := defaultFilePermissions
		switch {
		case generics.SliceContains(name, hooks.ValidHooks()):
			permissions = defaultFileExecPermissions
		case isBundledTerraformBinary(name):
			// Build runner vendors the terraform CLI as
			// .terraform-binaries/<os>_<arch>/terraform; OCI artifacts
			// don't preserve mode bits, so we have to re-apply the
			// exec bit on unpack or DetectBundledBinary will reject
			// the binary as non-executable. The VERSION sidecar in
			// the same dir is intentionally not exec.
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
