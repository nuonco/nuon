package emitter

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/go-playground/validator/v10"

	temporalclient "github.com/nuonco/nuon/pkg/temporal/client"
	"github.com/nuonco/nuon/services/ctl-api/internal"
)

type Params struct {
	fx.In

	Cfg     *internal.Config
	V       *validator.Validate
	DB      *gorm.DB `name:"psql"`
	TClient temporalclient.Client
	L       *zap.Logger
}

func NewWorkflows(params Params) (*Workflows, error) {
	return &Workflows{
		cfg:     params.Cfg,
		v:       params.V,
		db:      params.DB,
		tClient: params.TClient,
		l:       params.L,
	}, nil
}

type Workflows struct {
	cfg     *internal.Config
	v       *validator.Validate
	db      *gorm.DB
	tClient temporalclient.Client
	l       *zap.Logger
}

func (w *Workflows) All() []any {
	return []any{
		w.Emitter,
		w.CronTicker,
	}
}
