package ocicopy

import (
	"context"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"

	"github.com/nuonco/nuon/bins/runner/internal/pkg/oci"
	"github.com/nuonco/nuon/pkg/plugins/configs"
)

func (c *copier) CopyFromStore(ctx context.Context, store oras.ReadOnlyTarget, srcTag string, dstCfg *configs.OCIRegistryRepository, dstTag string) (*ocispec.Descriptor, error) {
	dstRepo, err := oci.GetRepo(ctx, dstCfg)
	if err != nil {
		return nil, err
	}

	res, err := oras.Copy(ctx, store, srcTag, dstRepo, dstTag, oras.DefaultCopyOptions)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
