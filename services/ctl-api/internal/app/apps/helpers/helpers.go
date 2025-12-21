package helpers

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/go-github/v50/github"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type Params struct {
	fx.In

	Cfg         *internal.Config
	GhClient    *github.Client
	DB          *gorm.DB `name:"psql"`
	V           *validator.Validate
	L           *zap.Logger
	VcsHelpers  *vcshelpers.Helpers
	QueueClient *queueclient.Client
}

type Helpers struct {
	cfg         *internal.Config
	ghClient    *github.Client
	db          *gorm.DB
	v           *validator.Validate
	l           *zap.Logger
	vcsHelpers  *vcshelpers.Helpers
	queueClient *queueclient.Client
}

func New(params Params) *Helpers {
	return &Helpers{
		v:           params.V,
		cfg:         params.Cfg,
		ghClient:    params.GhClient,
		db:          params.DB,
		l:           params.L,
		vcsHelpers:  params.VcsHelpers,
		queueClient: params.QueueClient,
	}
}

// VCSHelpers returns the VCS helpers for git source resolution and token creation.
func (h *Helpers) VCSHelpers() *vcshelpers.Helpers {
	return h.vcsHelpers
}
