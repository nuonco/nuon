package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	componenthelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/api"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/features"

	appshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/helpers"
)

type Params struct {
	fx.In

	V                *validator.Validate
	L                *zap.Logger
	DB               *gorm.DB `name:"psql"`
	MW               metrics.Writer
	Cfg              *internal.Config
	ComponentHelpers *componenthelpers.Helpers
	Helpers          *helpers.Helpers
	AppsHelpers      *appshelpers.Helpers
	FeaturesClient   *features.Features
	EvClient         eventloop.Client
}

type service struct {
	v                *validator.Validate
	l                *zap.Logger
	db               *gorm.DB
	mw               metrics.Writer
	cfg              *internal.Config
	componentHelpers *componenthelpers.Helpers
	helpers          *helpers.Helpers
	appsHelpers      *appshelpers.Helpers
	featuresClient   *features.Features
	evClient         eventloop.Client
}

var _ api.Service = (*service)(nil)

func (s *service) RegisterPublicRoutes(api *gin.Engine) error {
	// get all installs across orgs
	api.GET("/v1/installs", s.GetOrgInstalls)

	// get / create installs for an app
	api.GET("/v1/apps/:app_id/installs", s.GetAppInstalls)
	api.POST("/v1/apps/:app_id/installs", s.CreateInstall)

	// individual installs
	api.GET("/v1/installs/:install_id", s.GetInstall)
	api.PATCH("/v1/installs/:install_id", s.UpdateInstall)
	api.DELETE("/v1/installs/:install_id", s.DeleteInstall)
	api.POST("/v1/installs/:install_id/reprovision", s.ReprovisionInstall)

	api.POST("/v1/installs/:install_id/deprovision", s.DeprovisionInstall)
	api.POST("/v1/installs/:install_id/forget", s.ForgetInstall)
	api.POST("/v1/installs/:install_id/retry-workflow", s.RetryWorkflow) // Deprecated, use workflows instead

	// install deploys
	api.GET("/v1/installs/:install_id/deploys", s.GetInstallDeploys)
	api.POST("/v1/installs/:install_id/deploys", s.CreateInstallDeploy)
	api.GET("/v1/installs/:install_id/deploys/latest", s.GetInstallLatestDeploy)
	api.GET("/v1/installs/:install_id/deploys/:deploy_id", s.GetInstallDeploy)

	// install readme
	api.GET("/v1/installs/:install_id/readme", s.GetInstallReadme)

	// install state
	api.GET("/v1/installs/:install_id/state", s.GetInstallState)
	api.GET("/v1/installs/:install_id/state-history", s.GetInstallStateHistory)

	// install sandbox
	api.POST("/v1/installs/:install_id/reprovision-sandbox", s.ReprovisionInstallSandbox)
	api.POST("/v1/installs/:install_id/deprovision-sandbox", s.DeprovisionInstallSandbox)
	api.GET("/v1/installs/:install_id/sandbox-runs", s.GetInstallSandboxRuns)
	api.GET("/v1/installs/sandbox-runs/:run_id", s.GetInstallSandboxRun)

	// install inputs
	api.GET("/v1/installs/:install_id/inputs", s.GetInstallInputs)
	api.POST("/v1/installs/:install_id/inputs", s.CreateInstallInputs)
	api.GET("/v1/installs/:install_id/inputs/current", s.GetInstallCurrentInputs)
	api.PATCH("/v1/installs/:install_id/inputs", s.UpdateInstallInputs)

	// install components
	api.GET("/v1/installs/:install_id/components", s.GetInstallComponents)
	api.GET("/v1/installs/:install_id/components/summary", s.GetInstallComponentSummary)
	api.POST("/v1/installs/:install_id/components/teardown-all", s.TeardownInstallComponents)
	api.POST("/v1/installs/:install_id/components/deploy-all", s.DeployInstallComponents)
	api.GET("/v1/installs/:install_id/components/:component_id", s.GetInstallComponent)
	api.POST("/v1/installs/:install_id/components/:component_id/teardown", s.TeardownInstallComponent)
	api.GET("/v1/installs/:install_id/components/:component_id/deploys", s.GetInstallComponentDeploys)
	api.GET("/v1/installs/:install_id/components/:component_id/outputs", s.GetInstallComponentOutputs)
	api.GET("/v1/installs/:install_id/components/:component_id/deploys/latest", s.GetInstallComponentLatestDeploy)
	api.POST("/v1/installs/:install_id/sync-secrets", s.SyncSecrets)

	// install events
	api.GET("/v1/installs/:install_id/events", s.GetInstallEvents)
	api.GET("/v1/installs/:install_id/events/:event_id", s.GetInstallEvent)

	// workflows
	api.GET("/v1/installs/:install_id/workflows", s.GetWorkflows)
	api.GET("/v1/workflows/:workflow_id", s.GetWorkflow)
	api.PATCH("/v1/workflows/:workflow_id", s.UpdateWorkflow)
	api.GET("/v1/workflows/:workflow_id/steps", s.GetWorkflowSteps)
	api.GET("/v1/workflows/:workflow_id/steps/:workflow_step_id", s.GetWorkflowStep)
	api.POST("/v1/workflows/:workflow_id/cancel", s.CancelWorkflow)
	api.GET("/v1/workflows/:workflow_id/steps/:workflow_step_id/approvals/:approval_id", s.GetWorkflowStepApproval)
	api.POST("/v1/workflows/:workflow_id/steps/:workflow_step_id/approvals/:approval_id/response", s.CreateWorkflowStepApprovalResponse)
	api.GET("/v1/workflows/:workflow_id/steps/:workflow_step_id/approvals/:approval_id/contents", s.GetWorkflowStepApprovalContents)
	// retry workflow
	api.POST("/v1/workflows/:workflow_id/retry", s.RetryOwnerWorkflow)

	// deprecated
	api.GET("/v1/install-workflows/:install_workflow_id", s.GetInstallWorkflow)
	api.PATCH("/v1/install-workflows/:install_workflow_id", s.UpdateInstallWorkflow)
	api.GET("/v1/install-workflows/:install_workflow_id/steps", s.GetInstallWorkflowSteps)
	api.GET("/v1/install-workflows/:install_workflow_id/steps/:install_workflow_step_id", s.GetInstallWorkflowStep)
	api.POST("/v1/install-workflows/:install_workflow_id/cancel", s.CancelInstallWorkflow)
	api.GET("/v1/install-workflows/:install_workflow_id/steps/:install_workflow_step_id/approvals/:approval_id", s.GetInstallWorkflowStepApproval)
	api.POST("/v1/install-workflows/:install_workflow_id/steps/:install_workflow_step_id/approvals/:approval_id/response", s.CreateInstallWorkflowStepApprovalResponse)

	// install runner group
	api.GET("/v1/installs/:install_id/runner-group", s.GetInstallRunnerGroup)

	// phone home
	api.POST("/v1/installs/:install_id/phone-home/:phone_home_id", s.InstallPhoneHome)

	// install stacks
	api.GET("/v1/installs/:install_id/stack", s.GetInstallStackByInstallID)
	api.GET("/v1/installs/stacks/:stack_id", s.GetInstallStackByStackID)
	api.GET("/v1/installs/:install_id/stack-runs", s.GetInstallStackRuns)

	// install config
	api.POST("/v1/installs/:install_id/configs", s.CreateInstallConfig)
	api.PATCH("/v1/installs/:install_id/configs/:config_id", s.UpdateInstallConfig)

	// install audit logs
	api.GET("/v1/installs/:install_id/audit_logs", s.GetInstallAuditLogs)

	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	api.GET("/v1/installs", s.GetAllInstalls)
	api.POST("/v1/installs/admin-forget-account-installs", s.ForgetAccountInstalls)

	api.POST("/v1/installs/:install_id/admin-restart", s.RestartInstall)
	api.GET("/v1/installs/:install_id/admin-get", s.AdminGetInstall)
	api.GET("/v1/installs/:install_id/admin-get-runner-group", s.AdminGetInstallRunnerGroup)
	api.GET("/v1/installs/:install_id/admin-get-runner", s.AdminGetInstallRunner)
	api.POST("/v1/orgs/:org_id/admin-forget-installs", s.ForgetOrgInstalls)
	api.PATCH("/v1/installs/:install_id/admin-update-runner", s.AdminUpdateInstallRunner)
	api.POST("/v1/installs/:install_id/admin-generate-state", s.AdminInstallGenerateInstallState)
	api.GET("/v1/orgs/:org_id/admin-get-installs", s.AdminGetOrgInstalls)

	// NOTE(JM): the following endpoints should be removed after workflows/independent runners are rolled out
	api.POST("/v1/installs/:install_id/admin-reprovision", s.ReprovisionInstall)
	api.POST("/v1/installs/:install_id/admin-forget", s.AdminForgetInstall)
	api.POST("/v1/installs/:install_id/admin-update-sandbox", s.AdminUpdateSandbox)

	// temp for hackathon
	api.POST("/v1/admin-install-workflow-step-approve", s.AdminInstallWorkflowStepApprove)

	return nil
}

func (s *service) RegisterRunnerRoutes(api *gin.Engine) error {
	return nil
}

func New(params Params) *service {
	return &service{
		cfg:              params.Cfg,
		l:                params.L,
		v:                params.V,
		db:               params.DB,
		mw:               params.MW,
		componentHelpers: params.ComponentHelpers,
		helpers:          params.Helpers,
		evClient:         params.EvClient,
		appsHelpers:      params.AppsHelpers,
		featuresClient:   params.FeaturesClient,
	}
}
