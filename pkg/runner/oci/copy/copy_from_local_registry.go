package ocicopy

import (
	"context"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/registry/remote"

	"github.com/nuonco/nuon/pkg/plugins/configs"
	"github.com/nuonco/nuon/pkg/runner/oci"
	"github.com/nuonco/nuon/pkg/runner/op"
	"github.com/nuonco/nuon/pkg/runner/registry/local"
)

func (c *copier) CopyFromLocalRegistry(ctx context.Context, srcTag string, dstCfg *configs.OCIRegistryRepository, dstTag string) (_ *ocispec.Descriptor, retErr error) {
	opCtx, end := op.Tool(ctx, "oci", "copy_from_local_registry")
	ctx = opCtx
	defer func() { end(retErr) }()

	localRepo := local.GetCopyRepo(c.cfg)
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
