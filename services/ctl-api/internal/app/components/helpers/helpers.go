package helpers

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/go-github/v50/github"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	vcshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/vcs/helpers"
)

type Params struct {
	fx.In

	V          *validator.Validate
	Cfg        *internal.Config
	DB         *gorm.DB `name:"psql"`
	VcsHelpers *vcshelpers.Helpers
}

type Helpers struct {
	cfg        *internal.Config
	ghClient   *github.Client
	db         *gorm.DB
	vcsHelpers *vcshelpers.Helpers
}

func New(params Params) *Helpers {
	return &Helpers{
		cfg:        params.Cfg,
		db:         params.DB,
		vcsHelpers: params.VcsHelpers,
	}
}
