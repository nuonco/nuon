package emitter

import (
	"fmt"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/pkg/metrics"
	temporalclient "github.com/nuonco/nuon/pkg/temporal/client"
	tmetrics "github.com/nuonco/nuon/pkg/temporal/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
)

type Params struct {
	fx.In

	Cfg           *internal.Config
	V             *validator.Validate
	DB            *gorm.DB `name:"psql"`
	TClient       temporalclient.Client
	L             *zap.Logger
	MetricsWriter metrics.Writer
}

func NewWorkflows(params Params) (*Workflows, error) {
	tmw, err := tmetrics.New(params.V,
		tmetrics.WithMetricsWriter(params.MetricsWriter),
		tmetrics.WithTags(map[string]string{
			"context": "queue-emitter",
		}))
	if err != nil {
		return nil, fmt.Errorf("unable to create temporal metrics writer: %w", err)
	}
	return &Workflows{
		cfg:     params.Cfg,
		v:       params.V,
		db:      params.DB,
		tClient: params.TClient,
		l:       params.L,
		mw:      tmw,
	}, nil
}

type Workflows struct {
	cfg     *internal.Config
	v       *validator.Validate
	db      *gorm.DB
	tClient temporalclient.Client
	l       *zap.Logger
	mw      tmetrics.Writer
}

func (w *Workflows) All() []any {
	return []any{
		w.Emitter,
		w.CronTicker,
	}
}
