package ociarchive

import (
	"context"

	"github.com/spf13/afero"
	"go.uber.org/zap"
	"oras.land/oras-go/v2"

	"github.com/nuonco/nuon/pkg/plugins/configs"
	"github.com/nuonco/nuon/pkg/rootfs"
	"github.com/nuonco/nuon/pkg/sourcefs"
)

type Archive interface {
	Initialize(ctx context.Context) error
	Pack(ctx context.Context, log *zap.Logger, source *sourcefs.FS, filepaths []FileRef) error
	Unpack(ctx context.Context, repo *configs.OCIRegistryRepository, tag string) error
	Ref() oras.ReadOnlyTarget
	TmpDir() string
	Cleanup(context.Context) error
	BasePath() string
	FS() afero.Fs
}

var _ Archive = (*archive)(nil)

type archive struct {
	tmpDir   string
	chartDir string
	basePath string
	store    *archiveStore
	workFS   *rootfs.FS
}

func New() *archive {
	return &archive{}
}
