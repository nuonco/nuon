package ocicopy

import (
	"context"

	"github.com/go-playground/validator/v10"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"go.uber.org/fx"
	"oras.land/oras-go/v2"

	"github.com/powertoolsdev/mono/pkg/plugins/configs"
)

type Copier interface {
	Copy(ctx context.Context, srcCfg *configs.OCIRegistryRepository, srcTag string, dstCfg *configs.OCIRegistryRepository, dstTag string) (*ocispec.Descriptor, error)

	// store is useful for copying from a local store
	CopyFromStore(ctx context.Context, store oras.ReadOnlyTarget, srcTag string, dstCfg *configs.OCIRegistryRepository, dstTag string) (*ocispec.Descriptor, error)

	CopyFromLocalRegistry(ctx context.Context, localTag string, dstCfg *configs.OCIRegistryRepository, dstTag string) (*ocispec.Descriptor, error)
}

// this package supports the following types of inputs and outputs:
//
// Inputs
// public container image
// container image with encodedAuth
// oras.Store
//
// Outputs
// ECR repository with default credentials
// ECR repository with IAM role
// ACR repository with credentials baked in
//
// The only requirement is that an OCI registry config or ORAS store is passed in.
type copier struct {
	v *validator.Validate
}

var _ Copier = (*copier)(nil)

type CopierParams struct {
	fx.In

	V *validator.Validate
}

func New(params CopierParams) Copier {
	return &copier{
		v: params.V,
	}
}
