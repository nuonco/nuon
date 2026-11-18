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
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
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
	jwks                oidcissuer.JWKS
	helpers             *cloudconnectionshelpers.Helpers
	enqueueVerification func(context.Context, *app.CloudConnection) error
	now                 func() time.Time
}

var _ apiPkg.Service = (*service)(nil)

func New(params Params) (*service, error) {
	issuer, jwks, err := cloudconnections.IssuerFromConfig(params.Cfg)
	if err != nil {
		return nil, err
	}
	return &service{
		RouteRegister:       apiPkg.RouteRegister{EndpointAudit: params.EndpointAudit},
		db:                  params.DB,
		issuer:              issuer,
		jwks:                jwks,
		helpers:             params.Helpers,
		enqueueVerification: params.Helpers.EnqueueVerification,
		now:                 time.Now,
	}, nil
}

func (s *service) RegisterPublicRoutes(api *gin.Engine) error {
	routes := require.Routes(api.Group("/v1/cloud-connections"), permissions.APICtl)
	routes.GET("", require.Organization(permissions.CloudConnectionRead), s.List)
	routes.POST("", require.Organization(permissions.CloudConnectionCreate), s.Create)
	routes.GET("/:connection_id", require.Organization(permissions.CloudConnectionRead), s.Get)
	routes.DELETE("/:connection_id", require.Organization(permissions.CloudConnectionDelete), s.Delete)
	routes.POST("/:connection_id/verify", require.Organization(permissions.CloudConnectionVerify), s.Verify)
	routes.GET("/:connection_id/setup", require.Organization(permissions.CloudConnectionRead), s.Setup)

	repositories := require.Routes(api.Group("/v1/image-repositories"), permissions.APICtl)
	repositories.POST("", require.Organization(permissions.ImageRepositoryCreate), s.CreateImageRepository)
	repositories.GET("", require.Organization(permissions.ImageRepositoryRead), s.ListImageRepositories)
	repositories.GET("/:repository_id", require.Organization(permissions.ImageRepositoryRead), s.GetImageRepository)
	repositories.DELETE("/:repository_id", require.Organization(permissions.ImageRepositoryDelete), s.DeleteImageRepository)
	return nil
}

func (s *service) RegisterAuthRoutes(*gin.Engine) error           { return nil }
func (s *service) RegisterInternalRoutes(*gin.Engine) error       { return nil }
func (s *service) RegisterRunnerRoutes(*gin.Engine) error         { return nil }
func (s *service) RegisterAdminDashboardRoutes(*gin.Engine) error { return nil }
func (s *service) RegisterSlackRoutes(*gin.Engine) error          { return nil }
