package activities

import (
	"fmt"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/pkg/temporal/temporalzap"
	appshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/helpers"
	sharedactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/protos"
)

type Activities struct {
	db          *gorm.DB
	components  *protos.Adapter
	appsHelpers *appshelpers.Helpers
	evClient    eventloop.Client
	mw          metrics.Writer
	logger      *temporalzap.Logger

	*sharedactivities.Activities
}

type Params struct {
	fx.In

	DB          *gorm.DB `name:"psql"`
	Prt         *protos.Adapter
	AppsHelpers *appshelpers.Helpers
	SharedActs  *sharedactivities.Activities
	EvClient    eventloop.Client
	MW          metrics.Writer
}

func New(params Params) (*Activities, error) {
	logger, err := zap.NewProduction()
	tlogger := temporalzap.NewLogger(logger)
	if err != nil {
		return nil, fmt.Errorf("unable to create temporal logger: %w", err)
	}
	return &Activities{
		db:          params.DB,
		components:  params.Prt,
		appsHelpers: params.AppsHelpers,
		Activities:  params.SharedActs,
		evClient:    params.EvClient,
		mw:          params.MW,
		logger:      tlogger,
	}, nil
}
