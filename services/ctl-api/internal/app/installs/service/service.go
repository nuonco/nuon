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

	accountshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/accounts/helpers"
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
	AccountsHelpers  *accountshelpers.Helpers
	AppsHelpers      *appshelpers.Helpers
	FeaturesClient   *features.Features
	EvClient         eventloop.Client
	EndpointAudit    *api.EndpointAudit
}

type service struct {
	api.RouteRegister
	v                *validator.Validate
	l                *zap.Logger
	db               *gorm.DB
	mw               metrics.Writer
	cfg              *internal.Config
	componentHelpers *componenthelpers.Helpers
	helpers          *helpers.Helpers
	accountsHelpers  *accountshelpers.Helpers
	appsHelpers      *appshelpers.Helpers
	featuresClient   *features.Features
	evClient         eventloop.Client
}

var _ api.Service = (*service)(nil)

func (s *service) RegisterPublicRoutes(ge *gin.Engine) error {
	// get all installs across orgs
	ge.GET("/v1/installs", s.GetOrgInstalls)

	// get / create installs for an app
	ge.GET("/v1/apps/:app_id/installs", s.GetAppInstalls)
	ge.POST("/v1/apps/:app_id/installs", s.CreateInstall)

	// individual installs
	ge.GET("/v1/installs/:install_id", s.GetInstall)
	ge.PATCH("/v1/installs/:install_id", s.UpdateInstall)
	ge.DELETE("/v1/installs/:install_id", s.DeleteInstall)
	ge.POST("/v1/installs/:install_id/reprovision", s.ReprovisionInstall)

	ge.POST("/v1/installs/:install_id/deprovision", s.DeprovisionInstall)
	ge.POST("/v1/installs/:install_id/forget", s.ForgetInstall)
	ge.POST("/v1/installs/:install_id/retry-workflow", s.RetryWorkflow) // Deprecated, use workflows instead

	// install deploys
	ge.GET("/v1/installs/:install_id/deploys", s.GetInstallDeploys)
	ge.POST("/v1/installs/:install_id/deploys", s.CreateInstallDeploy)
	ge.GET("/v1/installs/:install_id/deploys/latest", s.GetInstallLatestDeploy)
	ge.GET("/v1/installs/:install_id/deploys/:deploy_id", s.GetInstallDeploy)

	// install readme
	ge.GET("/v1/installs/:install_id/readme", s.GetInstallReadme)

	// install drifts
	ge.GET("/v1/installs/:install_id/drifted-objects", s.GetDriftedObjects)

	// install state
	ge.GET("/v1/installs/:install_id/state", s.GetInstallState)
	ge.GET("/v1/installs/:install_id/state-history", s.GetInstallStateHistory)

	// install sandbox
	ge.POST("/v1/installs/:install_id/reprovision-sandbox", s.ReprovisionInstallSandbox)
	ge.POST("/v1/installs/:install_id/deprovision-sandbox", s.DeprovisionInstallSandbox)
	ge.GET("/v1/installs/:install_id/sandbox-runs", s.GetInstallSandboxRuns)
	ge.GET("/v1/installs/sandbox-runs/:run_id", s.GetInstallSandboxRun)

	// install inputs
	ge.GET("/v1/installs/:install_id/inputs", s.GetInstallInputs)
	ge.POST("/v1/installs/:install_id/inputs", s.CreateInstallInputs)
	ge.GET("/v1/installs/:install_id/inputs/current", s.GetInstallCurrentInputs)
	ge.PATCH("/v1/installs/:install_id/inputs", s.UpdateInstallInputs)

	// install components
	ge.GET("/v1/installs/:install_id/components", s.GetInstallComponents)
	ge.GET("/v1/installs/:install_id/components/summary", s.GetInstallComponentSummary)
	ge.POST("/v1/installs/:install_id/components/teardown-all", s.TeardownInstallComponents)
	ge.POST("/v1/installs/:install_id/components/deploy-all", s.DeployInstallComponents)
	ge.GET("/v1/installs/:install_id/components/:component_id", s.GetInstallComponent)
	ge.POST("/v1/installs/:install_id/components/:component_id/teardown", s.TeardownInstallComponent)
	ge.GET("/v1/installs/:install_id/components/:component_id/deploys", s.GetInstallComponentDeploys)
	ge.GET("/v1/installs/:install_id/components/:component_id/outputs", s.GetInstallComponentOutputs)
	ge.GET("/v1/installs/:install_id/components/:component_id/deploys/latest", s.GetInstallComponentLatestDeploy)
	ge.POST("/v1/installs/:install_id/sync-secrets", s.SyncSecrets)

	// install events
	ge.GET("/v1/installs/:install_id/events", s.GetInstallEvents)
	ge.GET("/v1/installs/:install_id/events/:event_id", s.GetInstallEvent)

	// workflows
	ge.GET("/v1/installs/:install_id/workflows", s.GetWorkflows)
	ge.GET("/v1/workflows/:workflow_id", s.GetWorkflow)
	ge.PATCH("/v1/workflows/:workflow_id", s.UpdateWorkflow)
	ge.GET("/v1/workflows/:workflow_id/steps", s.GetWorkflowSteps)
	ge.GET("/v1/workflows/:workflow_id/steps/:workflow_step_id", s.GetWorkflowStep)
	ge.POST("/v1/workflows/:workflow_id/cancel", s.CancelWorkflow)
	ge.GET("/v1/workflows/:workflow_id/steps/:workflow_step_id/approvals/:approval_id", s.GetWorkflowStepApproval)
	ge.POST("/v1/workflows/:workflow_id/steps/:workflow_step_id/approvals/:approval_id/response", s.CreateWorkflowStepApprovalResponse)
	ge.GET("/v1/workflows/:workflow_id/steps/:workflow_step_id/approvals/:approval_id/contents", s.GetWorkflowStepApprovalContents)
	// retry workflow
	ge.POST("/v1/workflows/:workflow_id/retry", s.RetryOwnerWorkflow)

	// deprecated
	s.GET(ge, "/v1/install-workflows/:install_workflow_id", s.GetInstallWorkflow, api.APIContextTypePublic, true)
	s.PATCH(ge, "/v1/install-workflows/:install_workflow_id", s.UpdateInstallWorkflow, api.APIContextTypePublic, true)
	s.GET(ge, "/v1/install-workflows/:install_workflow_id/steps", s.GetInstallWorkflowSteps, api.APIContextTypePublic, true)
	s.GET(ge, "/v1/install-workflows/:install_workflow_id/steps/:install_workflow_step_id", s.GetInstallWorkflowStep, api.APIContextTypePublic, true)
	s.POST(ge, "/v1/install-workflows/:install_workflow_id/cancel", s.CancelInstallWorkflow, api.APIContextTypePublic, true)
	s.GET(ge, "/v1/install-workflows/:install_workflow_id/steps/:install_workflow_step_id/approvals/:approval_id", s.GetInstallWorkflowStepApproval, api.APIContextTypePublic, true)
	// s.POST(ge, "/v1/install-workflows/:install_workflow_id/steps/:install_workflow_step_id/approvals/:approval_id/response", s.CreateInstallWorkflowStepApprovalResponse, api.APIContextTypePublic, true)

	// install runner group
	ge.GET("/v1/installs/:install_id/runner-group", s.GetInstallRunnerGroup)

	// phone home
	ge.POST("/v1/installs/:install_id/phone-home/:phone_home_id", s.InstallPhoneHome)

	// install stacks
	ge.GET("/v1/installs/:install_id/stack", s.GetInstallStackByInstallID)
	ge.GET("/v1/installs/stacks/:stack_id", s.GetInstallStackByStackID)
	ge.GET("/v1/installs/:install_id/stack-runs", s.GetInstallStackRuns)

	// install config
	ge.POST("/v1/installs/:install_id/configs", s.CreateInstallConfig)
	ge.PATCH("/v1/installs/:install_id/configs/:config_id", s.UpdateInstallConfig)

	// install audit logs
	ge.GET("/v1/installs/:install_id/audit_logs", s.GetInstallAuditLogs)

	// install cli config
	ge.GET("/v1/installs/:install_id/generate-cli-install-config", s.GenerateCLIInstallConfig)

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
		RouteRegister: api.RouteRegister{
			EndpointAudit: params.EndpointAudit,
		},
		cfg:              params.Cfg,
		l:                params.L,
		v:                params.V,
		db:               params.DB,
		mw:               params.MW,
		componentHelpers: params.ComponentHelpers,
		helpers:          params.Helpers,
		accountsHelpers:  params.AccountsHelpers,
		evClient:         params.EvClient,
		appsHelpers:      params.AppsHelpers,
		featuresClient:   params.FeaturesClient,
	}
}
