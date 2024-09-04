package helpers

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/go-github/v50/github"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
)

type Params struct {
	fx.In

	V        *validator.Validate
	Cfg      *internal.Config
	GhClient *github.Client
	DB       *gorm.DB `name:"psql"`
}

type Helpers struct {
	cfg      *internal.Config
	ghClient *github.Client
	db       *gorm.DB
}

func New(params Params) *Helpers {
	return &Helpers{
		cfg:      params.Cfg,
		ghClient: params.GhClient,
		db:       params.DB,
	}
}
