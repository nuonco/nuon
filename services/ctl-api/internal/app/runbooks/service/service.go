package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	runbookaccess "github.com/nuonco/nuon/services/ctl-api/internal/app/runbooks/access"
	runbookshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runbooks/helpers"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type Params struct {
	fx.In

	V              *validator.Validate
	DB             *gorm.DB `name:"psql"`
	Cfg            *internal.Config
	Helpers        *runbookshelpers.Helpers
	AppsHelpers    *appshelpers.Helpers
	InstallHelpers *installhelpers.Helpers
	EndpointAudit  *apiPkg.EndpointAudit
	QueueClient    *queueclient.Client
}

type service struct {
	apiPkg.RouteRegister
	v              *validator.Validate
	db             *gorm.DB
	cfg            *internal.Config
	helpers        *runbookshelpers.Helpers
	appsHelpers    *appshelpers.Helpers
	installHelpers *installhelpers.Helpers
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
		appsHelpers:    params.AppsHelpers,
		installHelpers: params.InstallHelpers,
		queueClient:    params.QueueClient,
	}
}

func (s *service) RegisterAuthRoutes(api *gin.Engine) error { return nil }

func (s *service) RegisterPublicRoutes(api *gin.Engine) error {
	apps := api.Group("/v1/apps/:app_id", require.AppInOrg(s.db))
	{
		runbooks := require.Routes(apps.Group("/runbooks"), permissions.APICtl)
		{
			runbooks.POST("", require.Point(permissions.RunbookCreate, s.runbookAppTarget), s.CreateRunbook)
			runbooks.GET("", s.runbookListRule(permissions.RunbookInspect, runbookaccess.Runbooks), s.GetRunbooks)
			runbooks.GET("/:runbook_id", require.Point(permissions.RunbookInspect, s.runbookReadTarget), s.GetRunbook)
			runbooks.PATCH("/:runbook_id", require.Point(permissions.RunbookUpdate, s.runbookTarget), s.UpdateRunbook)
			runbooks.DELETE("/:runbook_id", require.Point(permissions.RunbookDelete, s.runbookTarget), s.DeleteRunbook)
			runbooks.POST("/:runbook_id/configs", require.Point(permissions.RunbookConfigCreate, s.runbookConfigWriteTarget), s.CreateRunbookConfig)
			runbooks.GET("/:runbook_id/configs", s.runbookListRule(permissions.RunbookInspect, runbookaccess.Configs), s.GetRunbookConfigs)
		}
	}

	installs := api.Group("/v1/installs/:install_id")
	{
		installRunbooks := require.Routes(installs.Group("/runbooks"), permissions.APICtl)
		{
			installRunbooks.GET("", s.runbookListRule(permissions.InstallRunbookInspect, runbookaccess.InstallRunbooks), s.GetInstallRunbooks)
			installRunbooks.GET("/:runbook_id", require.Point(permissions.InstallRunbookInspect, s.installRunbookTarget), s.GetInstallRunbook)
			installRunbooks.POST("/:runbook_id/runs", require.Point(permissions.RunbookExecute, s.runbookExecutionTarget), s.CreateRunbookRun)
		}

		runbookRuns := require.Routes(installs.Group("/runbook-runs"), permissions.APICtl)
		{
			runbookRuns.GET("", s.runbookListRule(permissions.InstallRunbookInspect, runbookaccess.Runs), s.GetInstallRunbookRuns)
			runbookRuns.GET("/:run_id", require.Point(permissions.InstallRunbookInspect, s.runbookRunTarget), s.GetInstallRunbookRun)
		}
	}

	return nil
}
func (s *service) RegisterInternalRoutes(api *gin.Engine) error       { return nil }
func (s *service) RegisterRunnerRoutes(api *gin.Engine) error         { return nil }
func (s *service) RegisterAdminDashboardRoutes(api *gin.Engine) error { return nil }
func (s *service) RegisterSlackRoutes(api *gin.Engine) error          { return nil }
