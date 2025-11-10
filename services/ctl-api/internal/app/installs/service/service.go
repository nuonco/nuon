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
	runnershelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/helpers"
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
	RunnersHelpers   *runnershelpers.Helpers
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
	runnersHelpers   *runnershelpers.Helpers
	featuresClient   *features.Features
	evClient         eventloop.Client
}

var _ api.Service = (*service)(nil)

func (s *service) RegisterPublicRoutes(ge *gin.Engine) error {
	// get all installs across orgs
	ge.GET("/v1/installs", s.GetOrgInstalls)
	ge.POST("/v1/installs", s.CreateInstallV2)

	// get / create installs for an app
	apps := ge.Group("/v1/apps/:app_id")
	{
		apps.GET("/installs", s.GetAppInstalls)
		// apps.POST("/installs", s.CreateInstall)
		s.POST(apps, "/installs", s.CreateInstall, api.APIContextTypePublic, true) // Deprecated
	}

	// deprecated sandbox run route
	s.GET(ge, "/v1/installs/sandbox-runs/:run_id", s.GetInstallSandboxRun, api.APIContextTypePublic, true) // Deprecated

	// individual installs
	installs := ge.Group("/v1/installs/:install_id")
	{
		installs.GET("", s.GetInstall)
		installs.PATCH("", s.UpdateInstall)
		installs.DELETE("", s.DeleteInstall)
		installs.POST("/reprovision", s.ReprovisionInstall)
		installs.POST("/deprovision", s.DeprovisionInstall)
		installs.POST("/forget", s.ForgetInstall)
		s.POST(installs, "/retry-workflow", s.RetryWorkflow, api.APIContextTypePublic, true) // Deprecated

		// install deploys
		s.GET(installs, "/deploys", s.GetInstallDeploys, api.APIContextTypePublic, true)             // Deprecated
		s.POST(installs, "/deploys", s.CreateInstallDeploy, api.APIContextTypePublic, true)          // Deprecated
		s.GET(installs, "/deploys/latest", s.GetInstallLatestDeploy, api.APIContextTypePublic, true) // Deprecated
		s.GET(installs, "/deploys/:deploy_id", s.GetInstallDeploy, api.APIContextTypePublic, true)   // Deprecated
		installs.GET("/components/deploys", s.GetInstallComponentsDeploys)
		installs.GET("/components/:component_id/deploys/:deploy_id", s.GetInstallComponentDeploy)

		// install readme
		installs.GET("/readme", s.GetInstallReadme)

		// install drifts
		installs.GET("/drifted-objects", s.GetDriftedObjects)

		// install state
		installs.GET("/state", s.GetInstallState)
		installs.GET("/state-history", s.GetInstallStateHistory)

		// install sandbox
		installs.POST("/reprovision-sandbox", s.ReprovisionInstallSandbox)
		installs.POST("/deprovision-sandbox", s.DeprovisionInstallSandbox)

		sandboxRuns := installs.Group("/sandbox-runs")
		{
			sandboxRuns.GET("", s.GetInstallSandboxRuns)
			sandboxRuns.GET("/:run_id", s.GetInstallSandboxRunV2)
		}

		// install inputs
		inputs := installs.Group("/inputs")
		{
			inputs.GET("", s.GetInstallInputs)
			inputs.POST("", s.CreateInstallInputs)
			inputs.GET("/current", s.GetInstallCurrentInputs)
			inputs.PATCH("", s.UpdateInstallInputs)
		}

		// install components
		components := installs.Group("/components")
		{
			components.GET("", s.GetInstallComponents)
			components.GET("/summary", s.GetInstallComponentSummary)
			components.POST("/teardown-all", s.TeardownInstallComponents)
			components.POST("/deploy-all", s.DeployInstallComponents)

			component := components.Group("/:component_id")
			{
				component.GET("", s.GetInstallComponent)
				component.POST("/teardown", s.TeardownInstallComponent)
				component.GET("/deploys", s.GetInstallComponentDeploys)
				component.GET("/outputs", s.GetInstallComponentOutputs)
				component.GET("/deploys/latest", s.GetInstallComponentLatestDeploy)
				component.POST("/deploys", s.CreateInstallComponentDeploy)
			}
		}

		installs.POST("/sync-secrets", s.SyncSecrets)

		// install events
		events := installs.Group("/events")
		{
			events.GET("", s.GetInstallEvents)
			events.GET("/:event_id", s.GetInstallEvent)
		}

		// workflows for install
		installs.GET("/workflows", s.GetWorkflows)

		// install runner group
		installs.GET("/runner-group", s.GetInstallRunnerGroup)

		// phone home
		installs.POST("/phone-home/:phone_home_id", s.InstallPhoneHome)

		// install stacks
		installs.GET("/stack", s.GetInstallStackByInstallID)
		installs.GET("/stack-runs", s.GetInstallStackRuns)
		installs.GET("/generate-terraform-installer-config", s.GenerateTerraformInstallerConfig)

		// install config
		configs := installs.Group("/configs")
		{
			configs.POST("", s.CreateInstallConfig)
			configs.PATCH("/:config_id", s.UpdateInstallConfig)
		}

		// install audit logs
		installs.GET("/audit_logs", s.GetInstallAuditLogs)

		// install cli config
		installs.GET("/generate-cli-install-config", s.GenerateCLIInstallConfig)
	}

	// stack lookup by stack_id
	ge.GET("/v1/installs/stacks/:stack_id", s.GetInstallStackByStackID)

	// workflows (standalone)
	workflows := ge.Group("/v1/workflows/:workflow_id")
	{
		workflows.GET("", s.GetWorkflow)
		workflows.PATCH("", s.UpdateWorkflow)
		workflows.POST("/cancel", s.CancelWorkflow)
		s.POST(workflows, "/retry", s.RetryOwnerWorkflow, api.APIContextTypePublic, true) // Deprecated

		steps := workflows.Group("/steps")
		{
			steps.GET("", s.GetWorkflowSteps)
			steps.GET("/:step_id", s.GetWorkflowStep)
			steps.POST("/:step_id/retry", s.RetryWorkflowStep)

			approvals := steps.Group("/:step_id/approvals/:approval_id")
			{
				approvals.GET("", s.GetWorkflowStepApproval)
				approvals.POST("/response", s.CreateWorkflowStepApprovalResponse)
				approvals.GET("/contents", s.GetWorkflowStepApprovalContents)
			}
		}
	}

	// deprecated install-workflows

	s.GET(ge, "/v1/install-workflows/:install_workflow_id", s.GetInstallWorkflow, api.APIContextTypePublic, true)
	s.PATCH(ge, "/v1/install-workflows/:install_workflow_id", s.UpdateInstallWorkflow, api.APIContextTypePublic, true)
	s.GET(ge, "/v1/install-workflows/:install_workflow_id/steps", s.GetInstallWorkflowSteps, api.APIContextTypePublic, true)
	s.GET(ge, "/v1/install-workflows/:install_workflow_id/steps/:install_workflow_step_id", s.GetInstallWorkflowStep, api.APIContextTypePublic, true)
	s.POST(ge, "/v1/install-workflows/:install_workflow_id/cancel", s.CancelInstallWorkflow, api.APIContextTypePublic, true)
	s.GET(ge, "/v1/install-workflows/:install_workflow_id/steps/:install_workflow_step_id/approvals/:approval_id", s.GetInstallWorkflowStepApproval, api.APIContextTypePublic, true)

	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	// installs
	installs := api.Group("/v1/installs")
	{
		installs.GET("", s.GetAllInstalls)
		installs.POST("/admin-forget-account-installs", s.ForgetAccountInstalls)

		// install-specific admin routes
		install := installs.Group("/:install_id")
		{
			install.POST("/admin-restart", s.RestartInstall)
			install.GET("/admin-get", s.AdminGetInstall)
			install.GET("/admin-get-runner-group", s.AdminGetInstallRunnerGroup)
			install.GET("/admin-get-runner", s.AdminGetInstallRunner)
			install.PATCH("/admin-update-runner", s.AdminUpdateInstallRunner)
			install.POST("/admin-generate-state", s.AdminInstallGenerateInstallState)

			// NOTE(JM): the following endpoints should be removed after workflows/independent runners are rolled out
			install.POST("/admin-reprovision", s.ReprovisionInstall)
			install.POST("/admin-forget", s.AdminForgetInstall)
			install.POST("/admin-update-sandbox", s.AdminUpdateSandbox)
		}
	}

	// orgs
	orgs := api.Group("/v1/orgs/:org_id")
	{
		orgs.POST("/admin-forget-installs", s.ForgetOrgInstalls)
		orgs.GET("/admin-get-installs", s.AdminGetOrgInstalls)
	}

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
		runnersHelpers:   params.RunnersHelpers,
		featuresClient:   params.FeaturesClient,
	}
}
