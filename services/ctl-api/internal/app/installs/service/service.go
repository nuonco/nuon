package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	componenthelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/components/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/audit"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"

	accountshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/accounts/helpers"
	actionshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/actions/helpers"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	orgshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
	runnershelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
)

type Params struct {
	fx.In

	V                *validator.Validate
	L                *zap.Logger
	DB               *gorm.DB `name:"psql"`
	CHDB             *gorm.DB `name:"ch"`
	MW               metrics.Writer
	Cfg              *internal.Config
	ComponentHelpers *componenthelpers.Helpers
	Helpers          *helpers.Helpers
	AccountsHelpers  *accountshelpers.Helpers
	AppsHelpers      *appshelpers.Helpers
	OrgsHelpers      *orgshelpers.Helpers
	RunnersHelpers   *runnershelpers.Helpers
	ActionsHelpers   *actionshelpers.Helpers
	FeaturesClient   *features.Features
	QueueClient      *queueclient.Client
	FlowsClient      *flowclient.Client
	BlobService      blobstore.Service
	Audit            *audit.Emitter
	EndpointAudit    *api.EndpointAudit
}

type service struct {
	api.RouteRegister
	v                *validator.Validate
	l                *zap.Logger
	db               *gorm.DB
	chDB             *gorm.DB
	mw               metrics.Writer
	cfg              *internal.Config
	componentHelpers *componenthelpers.Helpers
	helpers          *helpers.Helpers
	accountsHelpers  *accountshelpers.Helpers
	appsHelpers      *appshelpers.Helpers
	orgsHelpers      *orgshelpers.Helpers
	runnersHelpers   *runnershelpers.Helpers
	actionsHelpers   *actionshelpers.Helpers
	featuresClient   *features.Features
	queueClient      *queueclient.Client
	flowsClient      *flowclient.Client
	blobSvc          blobstore.Service
	audit            *audit.Emitter
}

var _ api.Service = (*service)(nil)

func (s *service) RegisterPublicRoutes(ge *gin.Engine) error {
	// get all installs across orgs
	ge.GET("/v1/installs", s.GetOrgInstalls)
	ge.GET("/v1/installs/label-keys", s.GetInstallLabelKeys)
	ge.GET("/v1/installs/health", s.GetInstallsHealth)
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
		installs.POST("/labels", s.AddInstallLabels)
		installs.DELETE("/labels", s.RemoveInstallLabels)
		installs.POST("/reprovision", s.ReprovisionInstall)
		installs.POST("/reprovision-stack", s.ReprovisionInstallStack)
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

		// live component resource explorer
		installs.GET("/resources", s.GetInstallResources)

		// install-level component health rollup
		installs.GET("/health/timeline", s.GetInstallHealthTimeline)
		installs.POST("/health/baseline", s.ResetInstallHealthBaseline)
		installs.POST("/health/cluster-access", s.RefreshInstallHealthClusterAccess)

		// install state
		installs.GET("/state", s.GetInstallState)
		installs.GET("/state-history", s.GetInstallStateHistory)

		// install dns delegation check
		installs.GET("/dns/check", s.CheckInstallDNSDelegation)

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
			components.POST("/teardown-all", s.TeardownInstallComponents)
			components.POST("/deploy-all", s.DeployInstallComponents)

			component := components.Group("/:component_id")
			{
				component.GET("", s.GetInstallComponent)
				component.POST("/teardown", s.TeardownInstallComponent)
				component.POST("/toggle", s.ToggleInstallComponent)
				component.POST("/forget", s.ForgetInstallComponent)
				component.GET("/deploys", s.GetInstallComponentDeploys)
				component.GET("/outputs", s.GetInstallComponentOutputs)
				component.GET("/deploys/latest", s.GetInstallComponentLatestDeploy)
				component.POST("/deploys", s.CreateInstallComponentDeploy)

				// component health: gin can't mix wildcard names at the same path
				// depth, so these reuse the ":component_id" node above, but the
				// value they expect is the install_component's own ID (matching
				// the ClickHouse rows), not the catalog component ID the sibling
				// routes above use. The Swagger docs name it install_component_id.
				component.GET("/health/timeline", s.GetInstallComponentHealthTimeline)
				component.GET("/health/incident", s.GetInstallComponentHealthIncident)
				component.GET("/health/checks", s.GetInstallComponentHealthChecks)
				component.PUT("/health/checks/:check_name", s.PutInstallComponentHealthCheck)
			}
		}

		// install action workflows
		actions := installs.Group("/actions")
		{
			action := actions.Group("/:action_id")
			{
				action.GET("/outputs", s.GetInstallActionWorkflowOutputs)
			}
		}

		installs.POST("/sync-secrets", s.SyncSecrets)
		installs.POST("/sync-config", s.SyncInstallConfig)

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

		// runner bootstrap token
		installs.POST("/runner-bootstrap-token", s.CreateRunnerBootstrapToken)

		// install stacks
		installs.GET("/stack", s.GetInstallStackByInstallID)
		installs.GET("/stack-runs", s.GetInstallStackRuns)
		installs.GET("/generate-terraform-installer-config", s.GenerateTerraformInstallerConfig)

		// available roles
		installs.GET("/available-roles", s.GetAvailableRoles)

		// app permissions config with provisioning status
		installs.GET("/app-permissions-config", s.GetInstallAppPermissionsConfig)

		// install roles
		roles := installs.Group("/roles")
		{
			roles.GET("", s.GetInstallRoles)
			roles.GET("/latest", s.GetLatestInstallRoles)
			roles.GET("/usages", s.GetInstallRoleUsages)
			roles.PATCH("/:role_id", s.UpdateInstallRole)
		}

		// install config
		configs := installs.Group("/configs")
		{
			configs.POST("", s.CreateInstallConfig)
			configs.PATCH("/:config_id", s.UpdateInstallConfig)
		}

		// install app config versions
		installs.GET("/app-config-versions", s.GetInstallAppConfigVersions)
		installs.GET("/app-config-versions/:version_id/diff", s.GetInstallAppConfigVersionDiff)
		installs.GET("/config-versions", s.GetInstallConfigVersions)
		installs.GET("/config-versions/:version_id/diff", s.GetInstallConfigVersionDiff)
		installs.GET("/config-syncs", s.GetInstallConfigSyncs)
		installs.POST("/app-config-updates", s.CreateInstallAppConfigUpdate)

		// install audit logs
		installs.GET("/audit_logs", s.GetInstallAuditLogs)

		// install cli config
		installs.GET("/generate-cli-install-config", s.GenerateCLIInstallConfig)
	}

	// stack lookup by stack_id
	ge.GET("/v1/installs/stacks/:stack_id", s.GetInstallStackByStackID)

	// org-level workflow queries (must be registered before /:workflow_id group)
	ge.GET("/v1/workflows/pending-approvals", s.GetOrgPendingApprovals)
	ge.GET("/v1/workflows", s.GetOrgWorkflows)
	ge.POST("/v1/workflows/cancel", s.CancelWorkflows)

	// workflows (standalone, org-scoped)
	s.registerWorkflowSubtree(ge.Group("/v1/workflows/:workflow_id"))

	// workflows nested under install — ancestor-scoped: the workflow must be
	// owned by :install_id before any handler in the subtree runs.
	installWorkflows := ge.Group("/v1/installs/:install_id/workflows/:workflow_id")
	installWorkflows.Use(s.requireWorkflowOwner("installs", "install_id"))
	s.registerInstallWorkflowSubtree(installWorkflows)

	// workflows nested under app branch — same subtree, scoped to the branch owner.
	branchWorkflows := ge.Group("/v1/apps/:app_id/branches/:app_branch_id/workflows/:workflow_id")
	branchWorkflows.Use(s.requireWorkflowOwner("app_branches", "app_branch_id"))
	s.registerBranchWorkflowSubtree(branchWorkflows)

	// deprecated install-workflows

	s.GET(ge, "/v1/install-workflows/:install_workflow_id", s.GetWorkflowByInstall, api.APIContextTypePublic, true)
	s.PATCH(ge, "/v1/install-workflows/:install_workflow_id", s.UpdateWorkflowByInstall, api.APIContextTypePublic, true)
	s.GET(ge, "/v1/install-workflows/:install_workflow_id/steps", s.GetWorkflowStepsByInstall, api.APIContextTypePublic, true)
	s.GET(ge, "/v1/install-workflows/:install_workflow_id/steps/:install_workflow_step_id", s.GetWorkflowStepByInstall, api.APIContextTypePublic, true)
	s.POST(ge, "/v1/install-workflows/:install_workflow_id/cancel", s.CancelWorkflowByInstall, api.APIContextTypePublic, true)
	s.GET(ge, "/v1/install-workflows/:install_workflow_id/steps/:install_workflow_step_id/approvals/:approval_id", s.GetWorkflowStepApprovalByInstall, api.APIContextTypePublic, true)

	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	// installs
	installs := api.Group("/v1/installs")
	{
		installs.GET("", s.GetAllInstalls)
		installs.GET("/details", s.AdminListInstallsDetails)
		installs.POST("/admin-forget-account-installs", s.ForgetAccountInstalls)

		// install-specific admin routes
		install := installs.Group("/:install_id")
		{
			install.POST("/admin-restart", s.RestartInstall)
			install.POST("/admin-restart-queues", s.RestartInstallQueues)
			install.GET("/admin-get", s.AdminGetInstall)
			install.GET("/admin-get-runner-group", s.AdminGetInstallRunnerGroup)
			install.GET("/admin-get-runner", s.AdminGetInstallRunner)
			install.PATCH("/admin-update-runner", s.AdminUpdateInstallRunner)
			install.POST("/admin-reconcile-entities", s.AdminReconcileInstallEntities)
			install.POST("/admin-generate-state", s.AdminInstallGenerateInstallState)
			install.POST("/admin-generate-state-v2", s.AdminInstallGenerateInstallStateV2)

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

	// install stack version runs
	installStackVersionRuns := api.Group("/v1/install-stack-version-runs")
	{
		installStackVersionRun := installStackVersionRuns.Group("/:install_stack_version_run_id")
		{
			installStackVersionRun.POST("/admin-trigger-stack-output-update", s.AdminTriggerInstallStackOutputUpdate)
		}
	}

	// temp for hackathon
	api.POST("/v1/admin-install-workflow-step-approve", s.AdminInstallWorkflowStepApprove)

	return nil
}

func (s *service) RegisterRunnerRoutes(api *gin.Engine) error {
	// Read-only, side-effect-free config fetch consumed by the Terraform
	// provider's nuon_stack data source. Public: the per-stack-version
	// phone_home_id in the URL path is the secret.
	api.GET("/v1/stack-runs/:phone_home_id/config", s.GetInstallStackVersionConfig)

	// phone home reported by the install stack over the runner API. The
	// per-stack-version phone_home_id in the URL path is the secret; the route
	// is already in the public whitelist, so no runner token is required.
	api.POST("/v1/installs/:install_id/phone-home/:phone_home_id", s.InstallPhoneHome)
	return nil
}

func (s *service) RegisterAuthRoutes(api *gin.Engine) error {
	return nil
}

func (s *service) RegisterAdminDashboardRoutes(api *gin.Engine) error {
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
		chDB:             params.CHDB,
		mw:               params.MW,
		componentHelpers: params.ComponentHelpers,
		helpers:          params.Helpers,
		accountsHelpers:  params.AccountsHelpers,
		queueClient:      params.QueueClient,
		appsHelpers:      params.AppsHelpers,
		orgsHelpers:      params.OrgsHelpers,
		runnersHelpers:   params.RunnersHelpers,
		actionsHelpers:   params.ActionsHelpers,
		featuresClient:   params.FeaturesClient,
		flowsClient:      params.FlowsClient,
		blobSvc:          params.BlobService,
		audit:            params.Audit,
	}
}

func (s *service) RegisterSlackRoutes(api *gin.Engine) error {
	return nil
}

func (s *service) registerWorkflowSubtree(workflows *gin.RouterGroup) {
	workflows.GET("", s.GetWorkflow)
	workflows.PATCH("", s.UpdateWorkflow)
	workflows.POST("/cancel", s.CancelWorkflow)
	workflows.GET("/queue-position", s.GetWorkflowQueuePosition)

	stepGroups := workflows.Group("/step-groups")
	{
		stepGroups.GET("", s.GetWorkflowStepGroups)
		stepGroups.GET("/:step_group_id", s.GetWorkflowStepGroup)
	}

	steps := workflows.Group("/steps")
	{
		steps.GET("", s.GetWorkflowSteps)
		steps.GET("/:step_id", s.GetWorkflowStep)
		steps.GET("/:step_id/await", s.AwaitWorkflowStep)
		steps.POST("/:step_id/retry", s.RetryWorkflowStep)
		steps.POST("/:step_id/skip", s.SkipWorkflowStep)
		steps.POST("/:step_id/cancel", s.CancelWorkflowStep)

		approvals := steps.Group("/:step_id/approvals/:approval_id")
		{
			approvals.GET("", s.GetWorkflowStepApproval)
			approvals.POST("/response", s.CreateWorkflowStepApprovalResponse)
			approvals.GET("/contents", s.GetWorkflowStepApprovalContents)
		}
	}
}

func (s *service) requireWorkflowOwner(ownerType, ownerParam string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		orgID, err := cctx.OrgIDFromContext(ctx)
		if err != nil {
			ctx.Error(errors.Wrap(err, "unable to get org from context"))
			ctx.Abort()
			return
		}

		var count int64
		res := s.db.WithContext(ctx).Model(&app.Workflow{}).
			Where(app.Workflow{
				ID:        ctx.Param("workflow_id"),
				OrgID:     orgID,
				OwnerID:   ctx.Param(ownerParam),
				OwnerType: ownerType,
			}).
			Count(&count)
		if res.Error != nil {
			ctx.Error(errors.Wrap(res.Error, "unable to resolve workflow owner"))
			ctx.Abort()
			return
		}
		if count == 0 {
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "workflow not found"})
			return
		}

		ctx.Next()
	}
}

// Each nested route points at an annotated wrapper (see
// nested_install_workflows.go / nested_branch_workflows.go) so the path is a
// distinct swagger operation with a generated client method. The group guard
// is what makes the subtree ancestor-scoped.
func (s *service) registerInstallWorkflowSubtree(workflows *gin.RouterGroup) {
	workflows.GET("", s.GetWorkflowByInstall)
	workflows.PATCH("", s.UpdateWorkflowByInstall)
	workflows.POST("/cancel", s.CancelWorkflowByInstall)
	workflows.GET("/queue-position", s.GetWorkflowQueuePositionByInstall)

	stepGroups := workflows.Group("/step-groups")
	{
		stepGroups.GET("", s.GetWorkflowStepGroupsByInstall)
		stepGroups.GET("/:step_group_id", s.GetWorkflowStepGroupByInstall)
	}

	steps := workflows.Group("/steps")
	{
		steps.GET("", s.GetWorkflowStepsByInstall)
		steps.GET("/:step_id", s.GetWorkflowStepByInstall)
		steps.GET("/:step_id/await", s.AwaitWorkflowStepByInstall)
		steps.POST("/:step_id/retry", s.RetryWorkflowStepByInstall)
		steps.POST("/:step_id/skip", s.SkipWorkflowStepByInstall)
		steps.POST("/:step_id/cancel", s.CancelWorkflowStepByInstall)

		approvals := steps.Group("/:step_id/approvals/:approval_id")
		{
			approvals.GET("", s.GetWorkflowStepApprovalByInstall)
			approvals.POST("/response", s.CreateWorkflowStepApprovalResponseByInstall)
			approvals.GET("/contents", s.GetWorkflowStepApprovalContentsByInstall)
		}
	}
}

func (s *service) registerBranchWorkflowSubtree(workflows *gin.RouterGroup) {
	workflows.GET("", s.GetWorkflowByAppBranch)
	workflows.PATCH("", s.UpdateWorkflowByAppBranch)
	workflows.POST("/cancel", s.CancelWorkflowByAppBranch)
	workflows.GET("/queue-position", s.GetWorkflowQueuePositionByAppBranch)

	stepGroups := workflows.Group("/step-groups")
	{
		stepGroups.GET("", s.GetWorkflowStepGroupsByAppBranch)
		stepGroups.GET("/:step_group_id", s.GetWorkflowStepGroupByAppBranch)
	}

	steps := workflows.Group("/steps")
	{
		steps.GET("", s.GetWorkflowStepsByAppBranch)
		steps.GET("/:step_id", s.GetWorkflowStepByAppBranch)
		steps.GET("/:step_id/await", s.AwaitWorkflowStepByAppBranch)
		steps.POST("/:step_id/retry", s.RetryWorkflowStepByAppBranch)
		steps.POST("/:step_id/skip", s.SkipWorkflowStepByAppBranch)
		steps.POST("/:step_id/cancel", s.CancelWorkflowStepByAppBranch)

		approvals := steps.Group("/:step_id/approvals/:approval_id")
		{
			approvals.GET("", s.GetWorkflowStepApprovalByAppBranch)
			approvals.POST("/response", s.CreateWorkflowStepApprovalResponseByAppBranch)
			approvals.GET("/contents", s.GetWorkflowStepApprovalContentsByAppBranch)
		}
	}
}
