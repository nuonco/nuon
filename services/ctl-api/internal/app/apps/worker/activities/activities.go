package activities

import (
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/adapters/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/helpers"
)

type Activities struct {
	v       *validator.Validate
	db      *gorm.DB
	helpers *helpers.Helpers

	*activities.Activities
}

func New(cfg *internal.Config,
	v *validator.Validate,
	helpers *helpers.Helpers,
	sharedActs *activities.Activities,
	db *gorm.DB,
) (*Activities, error) {
	return &Activities{
		Activities: sharedActs,
		v:          v,
		db:         db,
		helpers:    helpers,
	}, nil
}
