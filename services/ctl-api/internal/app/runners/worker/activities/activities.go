package activities

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

type Params struct {
	fx.In

	Cfg              *internal.Config
	DB               *gorm.DB `name:"psql"`
	CHDB             *gorm.DB `name:"ch"`
	Helpers          *helpers.Helpers
	AuthzClient      *authz.Client
	AccountClient    *account.Client
	MW               metrics.Writer
	L                *zap.Logger
	QueueClient      *queueclient.Client
	StatusActivities *statusactivities.Activities
}

type Activities struct {
	db               *gorm.DB
	chDB             *gorm.DB
	helpers          *helpers.Helpers
	authzClient      *authz.Client
	acctClient       *account.Client
	cfg              *internal.Config
	mw               metrics.Writer
	l                *zap.Logger
	queueClient      *queueclient.Client
	statusActivities *statusactivities.Activities
}

func New(params Params) *Activities {
	return &Activities{
		cfg:              params.Cfg,
		db:               params.DB,
		chDB:             params.CHDB,
		helpers:          params.Helpers,
		authzClient:      params.AuthzClient,
		acctClient:       params.AccountClient,
		mw:               params.MW,
		l:                params.L,
		queueClient:      params.QueueClient,
		statusActivities: params.StatusActivities,
	}
}
