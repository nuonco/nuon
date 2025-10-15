package helpers

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/go-github/v50/github"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
)

type Params struct {
	fx.In

	Cfg      *internal.Config
	GhClient *github.Client
	DB       *gorm.DB `name:"psql"`
	V        *validator.Validate
	L        *zap.Logger
}

type Helpers struct {
	cfg      *internal.Config
	ghClient *github.Client
	db       *gorm.DB
	v        *validator.Validate
	l        *zap.Logger
}

func New(params Params) *Helpers {
	return &Helpers{
		v:        params.V,
		cfg:      params.Cfg,
		ghClient: params.GhClient,
		db:       params.DB,
    l: params.L,
	}
}
