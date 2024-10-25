package ocicopy

import (
	"context"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/registry/remote"

	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/oci"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/registry/local"
	"github.com/powertoolsdev/mono/pkg/plugins/configs"
)

func (c *copier) CopyFromLocalRegistry(ctx context.Context, srcTag string, dstCfg *configs.OCIRegistryRepository, dstTag string) (*ocispec.Descriptor, error) {
	localRepo := local.GetCopyRepo()
	repo, err := remote.NewRepository(localRepo)
	repo.PlainHTTP = true
	if err != nil {
		return nil, errors.Wrap(err, "unable to get local repo")
	}

	dstRepo, err := oci.GetRepo(ctx, dstCfg)
	if err != nil {
		return nil, err
	}

	res, err := oras.Copy(ctx, repo, srcTag, dstRepo, dstTag, oras.DefaultCopyOptions)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
