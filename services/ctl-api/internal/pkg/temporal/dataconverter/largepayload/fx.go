package largepayload

import (
	"go.temporal.io/sdk/converter"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
)

type Params struct {
	fx.In

	Cfg *internal.Config
	L   *zap.Logger
	DB  *gorm.DB `name:"psql"`
}

func New(params Params) converter.PayloadCodec {
	return &dataConverter{
		cfg: params.Cfg,
		l:   params.L,
		db:  params.DB,
	}
}

func AsLargePayload(f any) any {
	return fx.Annotate(
		f,
		fx.ResultTags(`name:"largepayload"`),
	)
}
