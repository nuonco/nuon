package activities

import (
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	temporalclient "github.com/nuonco/nuon/pkg/temporal/client"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/audit"
)

type Params struct {
	fx.In

	DB          *gorm.DB `name:"psql"`
	CHDB        *gorm.DB `name:"ch"`
	AppsHelpers *appshelpers.Helpers
	TClient     temporalclient.Client
	Cfg         *internal.Config
	Audit       *audit.Emitter `optional:"true"`
	L           *zap.Logger    `optional:"true"`

	MeterProvider metric.MeterProvider `optional:"true"`
}

type Activities struct {
	db          *gorm.DB
	chDB        *gorm.DB
	appsHelpers *appshelpers.Helpers
	tClient     temporalclient.Client
	cfg         *internal.Config
	audit       *audit.Emitter
	l           *zap.Logger

	policyEvaluationMetrics
	driftPlanEvaluations metric.Int64Counter
}

func New(params Params) *Activities {
	l := params.L
	if l == nil {
		l = zap.NewNop()
	}
	provider := params.MeterProvider
	if provider == nil {
		provider = noop.NewMeterProvider()
	}
	a := &Activities{
		db:          params.DB,
		chDB:        params.CHDB,
		appsHelpers: params.AppsHelpers,
		tClient:     params.TClient,
		cfg:         params.Cfg,
		audit:       params.Audit,
		l:           l,
	}
	a.policyEvaluationMetrics = newPolicyEvaluationMetrics(provider)
	a.driftPlanEvaluations, _ = provider.Meter("github.com/nuonco/nuon/ctl-api/drift-plan").Int64Counter(
		"nuon.install.drift.plan.evaluation.attempts",
		metric.WithUnit("{attempt}"),
		metric.WithDescription("Returned drift plan interpretation attempts, not end-to-end drift checks."))
	return a
}
