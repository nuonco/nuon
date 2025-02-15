package activities

import (
	"fmt"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/pkg/metrics"
	temporalclient "github.com/powertoolsdev/mono/pkg/temporal/client"
	"github.com/powertoolsdev/mono/pkg/temporal/temporalzap"
	appshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/protos"
)

type Activities struct {
	db          *gorm.DB
	chDB        *gorm.DB
	components  *protos.Adapter
	appsHelpers *appshelpers.Helpers
	evClient    eventloop.Client
	mw          metrics.Writer
	logger      *temporalzap.Logger
	tClient     temporalclient.Client
}

type Params struct {
	fx.In

	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	Prt            *protos.Adapter
	AppsHelpers    *appshelpers.Helpers
	EvClient       eventloop.Client
	MW             metrics.Writer
	TemporalClient temporalclient.Client
}

func New(params Params) (*Activities, error) {
	logger, err := zap.NewProduction()
	tlogger := temporalzap.NewLogger(logger)
	if err != nil {
		return nil, fmt.Errorf("unable to create temporal logger: %w", err)
	}
	return &Activities{
		db:          params.DB,
		chDB:        params.CHDB,
		components:  params.Prt,
		appsHelpers: params.AppsHelpers,
		evClient:    params.EvClient,
		mw:          params.MW,
		logger:      tlogger,
		tClient:     params.TemporalClient,
	}, nil
}
