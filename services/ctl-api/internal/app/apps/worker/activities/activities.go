package activities

import (
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/helpers"
	"gorm.io/gorm"
)

type Activities struct {
	v       *validator.Validate
	db      *gorm.DB
	helpers *helpers.Helpers
}

func New(cfg *internal.Config,
	v *validator.Validate,
	helpers *helpers.Helpers,
	db *gorm.DB) (*Activities, error) {
	return &Activities{
		v:       v,
		db:      db,
		helpers: helpers,
	}, nil
}
