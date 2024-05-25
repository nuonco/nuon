package activities

import (
	"gorm.io/gorm"

	sharedactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/protos"
	appshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/hooks"
)

type Activities struct {
	db          *gorm.DB
	components  *protos.Adapter
	appsHelpers *appshelpers.Helpers
	helpers     *helpers.Helpers
	hooks       *hooks.Hooks

	*sharedactivities.Activities
}

func New(db *gorm.DB,
	prt *protos.Adapter,
	appsHelpers *appshelpers.Helpers,
	helpers *helpers.Helpers,
	hooks *hooks.Hooks,
	sharedActs *sharedactivities.Activities,
) (*Activities, error) {
	return &Activities{
		db:          db,
		components:  prt,
		appsHelpers: appsHelpers,
		helpers:     helpers,
		hooks:       hooks,
		Activities:  sharedActs,
	}, nil
}
