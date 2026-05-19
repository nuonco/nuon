package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	runbookshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runbooks/helpers"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type Params struct {
	fx.In

	V              *validator.Validate
	DB             *gorm.DB `name:"psql"`
	Cfg            *internal.Config
	Helpers        *runbookshelpers.Helpers
	InstallHelpers *installhelpers.Helpers
	EndpointAudit  *apiPkg.EndpointAudit
	FeaturesClient *features.Features
	EvClient       eventloop.Client
	QueueClient    *queueclient.Client
}

type service struct {
	apiPkg.RouteRegister
	v              *validator.Validate
	db             *gorm.DB
	cfg            *internal.Config
	helpers        *runbookshelpers.Helpers
	installHelpers *installhelpers.Helpers
	featuresClient *features.Features
	evClient       eventloop.Client
	queueClient    *queueclient.Client
}

var _ apiPkg.Service = (*service)(nil)

func New(params Params) *service {
	return &service{
		RouteRegister: apiPkg.RouteRegister{
			EndpointAudit: params.EndpointAudit,
		},
		v:              params.V,
		db:             params.DB,
		cfg:            params.Cfg,
		helpers:        params.Helpers,
		installHelpers: params.InstallHelpers,
		featuresClient: params.FeaturesClient,
		evClient:       params.EvClient,
		queueClient:    params.QueueClient,
	}
}

func (s *service) RegisterAuthRoutes(api *gin.Engine) error { return nil }

func (s *service) RegisterPublicRoutes(api *gin.Engine) error {
	apps := api.Group("/v1/apps/:app_id")
	{
		runbooks := apps.Group("/runbooks")
		{
			runbooks.POST("", s.CreateRunbook)
			runbooks.GET("", s.GetRunbooks)
			runbooks.GET("/:runbook_id", s.GetRunbook)
			runbooks.PATCH("/:runbook_id", s.UpdateRunbook)
			runbooks.DELETE("/:runbook_id", s.DeleteRunbook)
			runbooks.POST("/:runbook_id/configs", s.CreateRunbookConfig)
			runbooks.GET("/:runbook_id/configs", s.GetRunbookConfigs)
		}
	}

	installs := api.Group("/v1/installs/:install_id")
	{
		installRunbooks := installs.Group("/runbooks")
		{
			installRunbooks.GET("", s.GetInstallRunbooks)
			installRunbooks.GET("/:runbook_id", s.GetInstallRunbook)
			installRunbooks.POST("/:runbook_id/runs", s.CreateRunbookRun)
		}

		runbookRuns := installs.Group("/runbook-runs")
		{
			runbookRuns.GET("", s.GetInstallRunbookRuns)
			runbookRuns.GET("/:run_id", s.GetInstallRunbookRun)
		}
	}

	return nil
}
func (s *service) RegisterInternalRoutes(api *gin.Engine) error       { return nil }
func (s *service) RegisterRunnerRoutes(api *gin.Engine) error         { return nil }
func (s *service) RegisterAdminDashboardRoutes(api *gin.Engine) error { return nil }
func (s *service) RegisterSlackRoutes(api *gin.Engine) error          { return nil }
