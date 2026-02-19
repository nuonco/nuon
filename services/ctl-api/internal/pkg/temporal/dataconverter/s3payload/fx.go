package s3payload

import (
	"go.temporal.io/sdk/converter"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
)

type Params struct {
	fx.In

	Cfg     *internal.Config
	L       *zap.Logger
	BlobSvc blobstore.Service
	MW      metrics.Writer
}

func New(params Params) converter.PayloadCodec {
	return &dataConverter{
		cfg:     params.Cfg,
		l:       params.L,
		blobSvc: params.BlobSvc,
		mw:      params.MW,
	}
}

func AsS3Payload(f any) any {
	return fx.Annotate(
		f,
		fx.ResultTags(`name:"s3payload"`),
	)
}
