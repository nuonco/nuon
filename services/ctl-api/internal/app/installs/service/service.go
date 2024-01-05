package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/pkg/waypoint/client/multi"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/adapters/terraformcloud"
	componenthelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/hooks"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type service struct {
	v                *validator.Validate
	l                *zap.Logger
	db               *gorm.DB
	mw               metrics.Writer
	cfg              *internal.Config
	hooks            *hooks.Hooks
	orgsOutputs      *terraformcloud.OrgsOutputs
	wpClient         multi.Client
	componentHelpers *componenthelpers.Helpers
}

func (s *service) RegisterRoutes(api *gin.Engine) error {
	// get all installs across orgs
	api.GET("/v1/installs", s.GetOrgInstalls)

	// get / create installs for an app
	api.GET("/v1/apps/:app_id/installs", s.GetAppInstalls)
	api.POST("/v1/apps/:app_id/installs", s.CreateInstall)

	// individual installs
	api.GET("/v1/installs/:install_id", s.GetInstall)
	api.PATCH("/v1/installs/:install_id", s.UpdateInstall)
	api.DELETE("/v1/installs/:install_id", s.DeleteInstall)

	// install deploys
	api.GET("/v1/installs/:install_id/deploys", s.GetInstallDeploys)
	api.POST("/v1/installs/:install_id/deploys", s.CreateInstallDeploy)
	api.GET("/v1/installs/:install_id/deploys/latest", s.GetInstallLatestDeploy)
	api.GET("/v1/installs/:install_id/deploys/:deploy_id", s.GetInstallDeploy)
	api.GET("/v1/installs/:install_id/deploys/:deploy_id/logs", s.GetInstallDeployLogs)
	api.GET("/v1/installs/:install_id/deploys/:deploy_id/plan", s.GetInstallDeployPlan)

	// install sandbox
	api.GET("/v1/installs/:install_id/sandbox-runs", s.GetInstallSandboxRuns)
	api.GET("/v1/installs/:install_id/sandbox-run/:run_id/logs", s.GetInstallSandboxRunLogs)

	// install inputs
	api.GET("/v1/installs/:install_id/inputs", s.GetInstallInputs)
	api.POST("/v1/installs/:install_id/inputs", s.CreateInstallInputs)
	api.GET("/v1/installs/:install_id/inputs/current", s.GetInstallCurrentInputs)

	// install components
	api.GET("/v1/installs/:install_id/components", s.GetInstallComponents)
	api.POST("/v1/installs/:install_id/components/:component_id/teardown", s.TeardownInstallComponent)
	api.GET("/v1/installs/:install_id/components/:component_id/deploys", s.GetInstallComponentDeploys)
	api.GET("/v1/installs/:install_id/components/:component_id/deploys/latest", s.GetInstallComponentLatestDeploy)

	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	api.GET("/v1/installs", s.GetAllInstalls)

	api.POST("/v1/installs/:install_id/admin-restart", s.RestartInstall)
	api.POST("/v1/installs/:install_id/admin-reprovision", s.ReprovisionInstall)
	api.POST("/v1/installs/:install_id/admin-deprovision", s.AdminDeprovisionInstall)
	api.POST("/v1/installs/:install_id/admin-delete", s.AdminDeleteInstall)
	api.POST("/v1/installs/:install_id/admin-forget", s.ForgetInstall)
	api.POST("/v1/installs/:install_id/admin-update-sandbox", s.AdminUpdateSandbox)
	api.POST("/v1/installs/:install_id/admin-teardown-components", s.AdminTeardownInstallComponents)
	api.POST("/v1/installs/:install_id/admin-deploy-components", s.AdminDeployInstallComponents)

	api.POST("/v1/orgs/:org_id/admin-forget-installs", s.ForgetOrgInstalls)
	return nil
}

func New(v *validator.Validate,
	cfg *internal.Config,
	db *gorm.DB,
	mw metrics.Writer,
	l *zap.Logger,
	hooks *hooks.Hooks,
	orgsOutputs *terraformcloud.OrgsOutputs,
	wpClient multi.Client,
	componentHelpers *componenthelpers.Helpers,
) *service {
	return &service{
		cfg:              cfg,
		l:                l,
		v:                v,
		db:               db,
		mw:               mw,
		hooks:            hooks,
		orgsOutputs:      orgsOutputs,
		wpClient:         wpClient,
		componentHelpers: componentHelpers,
	}
}
