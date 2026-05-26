package blob

import (
	"go.temporal.io/sdk/converter"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/filecache"
	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
)

type Params struct {
	fx.In

	Cfg     *internal.Config
	L       *zap.Logger
	DB      *gorm.DB `name:"psql"`
	BlobSvc blobstore.Service
	MW      metrics.Writer
	Cache   *filecache.FileCache
}

func New(params Params) converter.PayloadCodec {
	return &dataConverter{
		cfg:           params.Cfg,
		l:             params.L,
		db:            params.DB,
		blobSvc:       params.BlobSvc,
		mw:            params.MW,
		cache:         params.Cache,
		encodeEnabled: params.Cfg.LargePayloadType == "blob",
	}
}

func AsBlob(f any) any {
	return fx.Annotate(
		f,
		fx.ResultTags(`name:"blob"`),
	)
}
