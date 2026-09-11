package activities

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	temporalclient "github.com/nuonco/nuon/pkg/temporal/client"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/audit"
)

type WorkflowMetrics interface {
	FlowStarted(context.Context, app.Workflow)
}

type Params struct {
	fx.In

	DB          *gorm.DB `name:"psql"`
	CHDB        *gorm.DB `name:"ch"`
	AppsHelpers *appshelpers.Helpers
	TClient     temporalclient.Client
	Cfg         *internal.Config
	Audit       *audit.Emitter  `optional:"true"`
	L           *zap.Logger     `optional:"true"`
	Metrics     WorkflowMetrics `optional:"true"`
}

type Activities struct {
	db          *gorm.DB
	chDB        *gorm.DB
	appsHelpers *appshelpers.Helpers
	tClient     temporalclient.Client
	cfg         *internal.Config
	audit       *audit.Emitter
	l           *zap.Logger
	metrics     WorkflowMetrics
}

func New(params Params) *Activities {
	l := params.L
	if l == nil {
		l = zap.NewNop()
	}
	return &Activities{
		db:          params.DB,
		chDB:        params.CHDB,
		appsHelpers: params.AppsHelpers,
		tClient:     params.TClient,
		cfg:         params.Cfg,
		audit:       params.Audit,
		l:           l,
		metrics:     params.Metrics,
	}
}
