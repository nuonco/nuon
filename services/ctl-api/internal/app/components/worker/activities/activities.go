package activities

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/protos"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/hooks"
	"gorm.io/gorm"
)

type Activities struct {
	db      *gorm.DB
	protos  *protos.Adapter
	helpers *helpers.Helpers
	hooks   *hooks.Hooks
}

func New(prt *protos.Adapter,
	db *gorm.DB,
	helpers *helpers.Helpers,
	hooks *hooks.Hooks,
) (*Activities, error) {
	return &Activities{
		db:      db,
		protos:  prt,
		helpers: helpers,
		hooks:   hooks,
	}, nil
}
