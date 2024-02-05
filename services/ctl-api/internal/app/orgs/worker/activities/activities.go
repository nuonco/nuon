package activities

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	apphooks "github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/hooks"
	"gorm.io/gorm"
)

type Activities struct {
	db       *gorm.DB
	appHooks *apphooks.Hooks
}

func New(cfg *internal.Config,
	appHooks *apphooks.Hooks,
	db *gorm.DB) (*Activities, error) {
	return &Activities{
		db:       db,
		appHooks: appHooks,
	}, nil
}
