package ocicopy

import (
	"context"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"

	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/oci"
	"github.com/powertoolsdev/mono/pkg/plugins/configs"
)

const (
	localRegistryRepositoryName string = "runner"
	localRegistryAddress        string = "http://localhost:5000"
)

func (c *copier) CopyFromLocalRegistry(ctx context.Context, srcTag string, dstCfg *configs.OCIRegistryRepository, dstTag string) (*ocispec.Descriptor, error) {
	srcCfg := &configs.OCIRegistryRepository{
		RegistryType: configs.OCIRegistryTypePublicOCI,
		Repository:   localRegistryRepositoryName,
		LoginServer:  localRegistryAddress,
	}

	srcRepo, err := oci.GetRepo(ctx, srcCfg)
	if err != nil {
		return nil, err
	}

	dstRepo, err := oci.GetRepo(ctx, dstCfg)
	if err != nil {
		return nil, err
	}

	res, err := oras.Copy(ctx, srcRepo, srcTag, dstRepo, dstTag, oras.DefaultCopyOptions)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
