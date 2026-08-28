package ociarchive

import (
	"context"

	"github.com/nuonco/nuon/pkg/plugins/configs"
	"go.uber.org/zap"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
)

type Archive interface {
	Initialize(ctx context.Context) error
	Pack(ctx context.Context, log *zap.Logger, filepaths []FileRef) error
	Unpack(ctx context.Context, repo *configs.OCIRegistryRepository, tag string) error
	UnpackFromStore(ctx context.Context, src oras.ReadOnlyTarget, ref string) error
	Ref() oras.ReadOnlyTarget
	TmpDir() string
	Cleanup(context.Context) error
	BasePath() string
}

// Source resolves a plan's OCI source tag to an artifact packaged in a local
// store, letting customer-managed runs unpack archive sources without any registry
// access. A false return means the tag is not packaged; callers must fail
// rather than fall back to the network.
type Source interface {
	ResolveArchive(tag string) (src oras.ReadOnlyTarget, ref string, ok bool)
}

var _ Archive = (*archive)(nil)

type archive struct {
	tmpDir   string
	chartDir string
	basePath string
	store    *file.Store
}

func New() *archive {
	return &archive{}
}
