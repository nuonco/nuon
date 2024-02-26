package activities

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/helpers"
	"gorm.io/gorm"
)

type Activities struct {
	db      *gorm.DB
	helpers *helpers.Helpers
}

func New(cfg *internal.Config,
	helpers *helpers.Helpers,
	db *gorm.DB) (*Activities, error) {
	return &Activities{
		db:      db,
		helpers: helpers,
	}, nil
}
