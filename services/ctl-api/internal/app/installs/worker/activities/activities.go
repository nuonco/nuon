package activities

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal/adapters/protos"
	appshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/hooks"
	"gorm.io/gorm"
)

type Activities struct {
	db          *gorm.DB
	components  *protos.Adapter
	appsHelpers *appshelpers.Helpers
	hooks       *hooks.Hooks
}

func New(db *gorm.DB,
	prt *protos.Adapter,
	appsHelpers *appshelpers.Helpers,
	hooks *hooks.Hooks,
) (*Activities, error) {
	return &Activities{
		db:          db,
		components:  prt,
		appsHelpers: appsHelpers,
		hooks:       hooks,
	}, nil
}
