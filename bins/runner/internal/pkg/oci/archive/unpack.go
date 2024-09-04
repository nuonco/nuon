package ociarchive

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/oci"
	"github.com/powertoolsdev/mono/pkg/plugins/configs"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
)

func (a *archive) Unpack(ctx context.Context, srcCfg *configs.OCIRegistryRepository, tag string) error {
	srcRepo, err := oci.GetRepo(ctx, srcCfg)
	if err != nil {
		return fmt.Errorf("unable to get source repo: %w", err)
	}

	manifest, err := oras.Copy(ctx, srcRepo, tag, a.store, tag, oras.DefaultCopyOptions)
	if err != nil {
		return fmt.Errorf("unable to copy image: %w", err)
	}

	_, err = content.FetchAll(ctx, a.store, manifest)
	if err != nil {
		return fmt.Errorf("unable to fetch contents: %w", err)
	}

	return nil
}
