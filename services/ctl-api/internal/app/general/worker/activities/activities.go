package activities

import (
	"fmt"

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

func New(db *gorm.DB,
	prt *protos.Adapter,
	appsHelpers *appshelpers.Helpers,
	sharedActs *sharedactivities.Activities,
	evClient eventloop.Client,
	mw metrics.Writer,
) (*Activities, error) {

	logger, err := zap.NewProduction()
	tlogger := temporalzap.NewLogger(logger)
	if err != nil {
		return nil, fmt.Errorf("unable to create temporal logger: %w", err)
	}
	return &Activities{
		db:          db,
		components:  prt,
		appsHelpers: appsHelpers,
		Activities:  sharedActs,
		evClient:    evClient,
		mw:          mw,
		logger:      tlogger,
	}, nil
}
