package activities

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/hooks"
	"gorm.io/gorm"
)

type Activities struct {
	db           *gorm.DB
	installHooks *hooks.Hooks
}

func New(cfg *internal.Config,
	installHooks *hooks.Hooks,
	db *gorm.DB) (*Activities, error) {
	return &Activities{
		db:           db,
		installHooks: installHooks,
	}, nil
}
