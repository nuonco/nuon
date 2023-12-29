package helpers

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/go-github/v50/github"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	vcshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/vcs/helpers"
	"gorm.io/gorm"
)

type Helpers struct {
	cfg        *internal.Config
	ghClient   *github.Client
	db         *gorm.DB
	vcsHelpers *vcshelpers.Helpers
}

func New(v *validator.Validate,
	cfg *internal.Config,
	db *gorm.DB,
	vcsHelpers *vcshelpers.Helpers,
) *Helpers {
	return &Helpers{
		cfg:        cfg,
		db:         db,
		vcsHelpers: vcsHelpers,
	}
}
