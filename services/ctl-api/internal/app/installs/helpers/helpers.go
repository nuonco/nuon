package helpers

import (
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	componenthelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/helpers"
	"gorm.io/gorm"
)

type Helpers struct {
	cfg              *internal.Config
	componentHelpers *componenthelpers.Helpers
	db               *gorm.DB
}

func New(v *validator.Validate,
	cfg *internal.Config,
	db *gorm.DB,
	componentHelpers *componenthelpers.Helpers,
) *Helpers {
	return &Helpers{
		cfg:              cfg,
		componentHelpers: componentHelpers,
		db:               db,
	}
}
