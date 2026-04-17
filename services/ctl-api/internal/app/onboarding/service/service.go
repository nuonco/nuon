package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	accountshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/accounts/helpers"
	installshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type service struct {
	apiPkg.RouteRegister
	v               *validator.Validate
	l               *zap.Logger
	db              *gorm.DB
	cfg             *internal.Config
	queueClient     *queueclient.Client
	installsHelpers *installshelpers.Helpers
	evClient        eventloop.Client
	featuresClient  *features.Features
	accountsHelpers *accountshelpers.Helpers
	catalog         *Catalog
}

var _ apiPkg.Service = (*service)(nil)

type Params struct {
	fx.In

	V               *validator.Validate
	L               *zap.Logger
	DB              *gorm.DB `name:"psql"`
	Cfg             *internal.Config
	EndpointAudit   *apiPkg.EndpointAudit
	QueueClient     *queueclient.Client
	InstallsHelpers *installshelpers.Helpers
	EvClient        eventloop.Client
	FeaturesClient  *features.Features
	AccountsHelpers *accountshelpers.Helpers
	Catalog         *Catalog
}

func New(params Params) *service {
	return &service{
		RouteRegister: apiPkg.RouteRegister{
			EndpointAudit: params.EndpointAudit,
		},
		v:               params.V,
		l:               params.L,
		db:              params.DB,
		cfg:             params.Cfg,
		queueClient:     params.QueueClient,
		installsHelpers: params.InstallsHelpers,
		evClient:        params.EvClient,
		featuresClient:  params.FeaturesClient,
		accountsHelpers: params.AccountsHelpers,
		catalog:         params.Catalog,
	}
}

func (s *service) RegisterPublicRoutes(api *gin.Engine) error {
	onboarding := api.Group("/v1/onboarding")
	{
		onboarding.GET("/example-apps", s.GetExampleApps)
		onboarding.POST("", s.CreateOnboarding)
		onboarding.GET("/current", s.GetCurrentOnboarding)
		onboarding.POST("/current/steps/organization", s.CompleteOrganizationStep)
		onboarding.POST("/current/steps/your-stack", s.CompleteYourStackStep)
		onboarding.POST("/current/steps/install", s.CompleteInstallStep)
		onboarding.POST("/current/steps/deploy", s.CompleteDeployStep)
		onboarding.POST("/current/steps/get-started", s.CompleteGetStartedStep)
		onboarding.DELETE("/current", s.AbandonOnboarding)
	}

	return nil
}

func (s *service) RegisterAuthRoutes(api *gin.Engine) error {
	return nil
}

func (s *service) RegisterRunnerRoutes(api *gin.Engine) error {
	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	return nil
}

func (s *service) RegisterAdminDashboardRoutes(api *gin.Engine) error {
	return nil
}
