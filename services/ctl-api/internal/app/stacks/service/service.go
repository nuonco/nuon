// Package service serves the runner API's `stacks` namespace: the authenticated
// endpoints an install stack uses to read its own configuration.
//
// Separate from the installs service, and from the older
// /v1/stack-runs/{phone_home_id}/config endpoint it supersedes, because those routes
// are public — the phone_home_id in the URL path is the only secret. Everything here
// authenticates normally, so the stack's service-account token (or an OIDC-federated
// token) identifies the caller and the standard org middleware scopes it.
package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	installshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
)

type Params struct {
	fx.In

	V               *validator.Validate
	DB              *gorm.DB `name:"psql"`
	MW              metrics.Writer
	L               *zap.Logger
	Cfg             *internal.Config
	EndpointAudit   *api.EndpointAudit
	InstallsHelpers *installshelpers.Helpers
	AcctClient      *account.Client
}

type service struct {
	api.RouteRegister
	v               *validator.Validate
	db              *gorm.DB
	mw              metrics.Writer
	l               *zap.Logger
	cfg             *internal.Config
	installsHelpers *installshelpers.Helpers
	acctClient      *account.Client
}

var _ api.Service = (*service)(nil)

func (s *service) RegisterRunnerRoutes(ge *gin.Engine) error {
	stacks := ge.Group("/v1/stacks/:install_id")
	{
		stacks.GET("/config", s.GetStackConfig)
	}

	return nil
}

// The token read is on the public API, not the runner API: it is the dashboard that
// displays it, and the customer applying the Terraform has no credential yet.
func (s *service) RegisterPublicRoutes(ge *gin.Engine) error {
	stacks := ge.Group("/v1/stacks/:install_id")
	{
		stacks.GET("/token", s.GetStackToken)
	}

	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	return nil
}

func (s *service) RegisterAuthRoutes(api *gin.Engine) error {
	return nil
}

func (s *service) RegisterAdminDashboardRoutes(api *gin.Engine) error {
	return nil
}

func (s *service) RegisterSlackRoutes(api *gin.Engine) error {
	return nil
}

func New(params Params) *service {
	return &service{
		RouteRegister: api.RouteRegister{
			EndpointAudit: params.EndpointAudit,
		},
		v:               params.V,
		db:              params.DB,
		mw:              params.MW,
		l:               params.L,
		cfg:             params.Cfg,
		installsHelpers: params.InstallsHelpers,
		acctClient:      params.AcctClient,
	}
}
