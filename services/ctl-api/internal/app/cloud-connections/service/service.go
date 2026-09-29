package service

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	cloudconnections "github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections"
	cloudconnectionshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections/helpers"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidcissuer"
)

type Params struct {
	fx.In

	DB            *gorm.DB `name:"psql"`
	Cfg           *internal.Config
	EndpointAudit *apiPkg.EndpointAudit
	Helpers       *cloudconnectionshelpers.Helpers
}

type service struct {
	apiPkg.RouteRegister
	db                  *gorm.DB
	issuer              *oidcissuer.Issuer
	helpers             *cloudconnectionshelpers.Helpers
	enqueueVerification func(context.Context, *app.CloudConnection) error
	now                 func() time.Time
}

var _ apiPkg.Service = (*service)(nil)

func New(params Params) (*service, error) {
	issuer, err := cloudconnections.IssuerFromConfig(params.Cfg)
	if err != nil {
		return nil, err
	}
	return &service{
		RouteRegister:       apiPkg.RouteRegister{EndpointAudit: params.EndpointAudit},
		db:                  params.DB,
		issuer:              issuer,
		helpers:             params.Helpers,
		enqueueVerification: params.Helpers.EnqueueVerification,
		now:                 time.Now,
	}, nil
}

func (s *service) RegisterPublicRoutes(api *gin.Engine) error {
	routes := api.Group("/v1/cloud-connections")
	routes.GET("", s.List)
	routes.POST("", s.Create)
	routes.GET("/:connection_id", s.Get)
	routes.DELETE("/:connection_id", s.Delete)
	routes.POST("/:connection_id/verify", s.Verify)
	routes.GET("/:connection_id/setup", s.Setup)
	return nil
}

func (s *service) RegisterAuthRoutes(*gin.Engine) error           { return nil }
func (s *service) RegisterInternalRoutes(*gin.Engine) error       { return nil }
func (s *service) RegisterRunnerRoutes(*gin.Engine) error         { return nil }
func (s *service) RegisterAdminDashboardRoutes(*gin.Engine) error { return nil }
func (s *service) RegisterSlackRoutes(*gin.Engine) error          { return nil }
