package activities

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/activities"
)

type Params struct {
	fx.In

	V          *validator.Validate
	Helpers    *helpers.Helpers
	SharedActs *activities.Activities
	DB         *gorm.DB `name:"psql"`
}

type Activities struct {
	v       *validator.Validate
	db      *gorm.DB
	helpers *helpers.Helpers

	*activities.Activities
}

func New(params Params) (*Activities, error) {
	return &Activities{
		Activities: params.SharedActs,
		v:          params.V,
		db:         params.DB,
		helpers:    params.Helpers,
	}, nil
}
