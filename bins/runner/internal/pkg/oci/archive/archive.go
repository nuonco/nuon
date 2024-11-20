package ociarchive

import (
	"context"

	"github.com/powertoolsdev/mono/pkg/plugins/configs"
	"go.uber.org/zap"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
)

type Archive interface {
	Initialize(ctx context.Context) error
	Pack(ctx context.Context, log *zap.Logger, filepaths []FileRef) error
	Unpack(ctx context.Context, repo *configs.OCIRegistryRepository, tag string) error
	Ref() oras.ReadOnlyTarget
	TmpDir() string
	Cleanup(context.Context) error
	BasePath() string
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
