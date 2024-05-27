package activities

import (
	"gorm.io/gorm"

	appshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/helpers"
	sharedactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/protos"
)

type Activities struct {
	db          *gorm.DB
	components  *protos.Adapter
	appsHelpers *appshelpers.Helpers
	helpers     *helpers.Helpers
	evClient    eventloop.Client

	*sharedactivities.Activities
}

func New(db *gorm.DB,
	prt *protos.Adapter,
	appsHelpers *appshelpers.Helpers,
	helpers *helpers.Helpers,
	sharedActs *sharedactivities.Activities,
	evClient eventloop.Client,
) (*Activities, error) {
	return &Activities{
		db:          db,
		components:  prt,
		appsHelpers: appsHelpers,
		helpers:     helpers,
		Activities:  sharedActs,
		evClient:    evClient,
	}, nil
}
