package gzip

import (
	"go.temporal.io/sdk/converter"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
)

type Params struct {
	fx.In

	Cfg *internal.Config
	L   *zap.Logger
	MW  metrics.Writer
}

func New(params Params) converter.PayloadCodec {
	return &dataConverter{
		cfg: params.Cfg,
		l:   params.L,
		mw:  params.MW,
	}
}

func AsGzip(f any) any {
	return fx.Annotate(
		f,
		fx.ResultTags(`name:"gzip"`),
	)
}
