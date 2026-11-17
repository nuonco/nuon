package service

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	componenthelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/components/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers/alignment"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/audit"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/telemetrytoken"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workloadjwt"

	accountshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/accounts/helpers"
	actionshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/actions/helpers"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	orgshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
	runnershelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
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
	AcctClient       *account.Client
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
	v                    *validator.Validate
	l                    *zap.Logger
	db                   *gorm.DB
	chDB                 *gorm.DB
	mw                   metrics.Writer
	cfg                  *internal.Config
	componentHelpers     *componenthelpers.Helpers
	helpers              *helpers.Helpers
	alignment            *alignment.Helpers
	accountsHelpers      *accountshelpers.Helpers
	acctClient           *account.Client
	appsHelpers          *appshelpers.Helpers
	orgsHelpers          *orgshelpers.Helpers
	runnersHelpers       *runnershelpers.Helpers
	actionsHelpers       *actionshelpers.Helpers
	featuresClient       *features.Features
	queueClient          *queueclient.Client
	flowsClient          *flowclient.Client
	blobSvc              blobstore.Service
	audit                *audit.Emitter
	workloadJWT          *workloadjwt.Verifier
	telemetryTokenIssuer *telemetrytoken.Issuer
}

var _ api.Service = (*service)(nil)

func (s *service) RegisterPublicRoutes(ge *gin.Engine) error {
	s.POST(ge, "/v1/installs/:install_id/phone-home/:phone_home_id", s.InstallPhoneHome, api.APIContextTypePublic, true)

	org := require.OrgRoute(ge)
	// get all installs across orgs
	orgAccess := require.Routes(org, permissions.APICtl)
	orgAccess.GET("/v1/installs", require.Collection(permissions.InstallInspect, s.installResources), s.GetOrgInstalls)
	orgAccess.GET("/v1/installs/label-keys", require.Collection(permissions.InstallInspect, s.installResources), s.GetInstallLabelKeys)
	orgAccess.GET("/v1/installs/health", require.Collection(permissions.InstallStateRead, s.installResources), s.GetInstallsHealth)
	orgAccess.POST("/v1/installs", require.Point(permissions.InstallCreate, s.createInstallTarget), s.CreateInstallV2)

	// get / create installs for an app
	apps := require.Routes(require.OrgRoute(ge).Group("/v1/apps/:app_id", require.AppInOrg(s.db)), permissions.APICtl)
	{
		apps.GET("/installs", require.Collection(permissions.InstallInspect, s.installResources), s.GetAppInstalls)
		s.POST(apps.For(require.Point(permissions.InstallCreate, s.createInstallTarget)), "/installs", s.CreateInstall, api.APIContextTypePublic, true) // Deprecated
	}

	// deprecated sandbox run route
	s.GET(orgAccess.For(require.Point(permissions.InstallHistoryRead, s.sandboxRunInstallTarget)), "/v1/installs/sandbox-runs/:run_id", s.GetInstallSandboxRun, api.APIContextTypePublic, true) // Deprecated

	installs := require.ResourceRoute(ge, permissions.KindInstall, "install_id").Group("/v1/installs/:install_id")
	installsAccess := require.Routes(installs, permissions.APICtl)
	{
		installsAccess.GET("", require.Point(permissions.InstallInspect, s.installTarget), s.GetInstall)
		installsAccess.PATCH("", require.Point(permissions.InstallUpdate, s.installIDTarget), s.UpdateInstall)
		installsAccess.DELETE("", require.Point(permissions.InstallDelete, s.installTarget), s.DeleteInstall)
		installsAccess.POST("/labels", require.Point(permissions.InstallUpdate, s.installIDTarget), s.AddInstallLabels)
		installsAccess.DELETE("/labels", require.Point(permissions.InstallUpdate, s.installIDTarget), s.RemoveInstallLabels)
		installsAccess.PATCH("/app-branch", require.Point(permissions.InstallConfigUpdate, s.installIDTarget), s.MoveInstallToAppBranch)
		installsAccess.POST("/reprovision", require.Point(permissions.InstallProvision, s.installLifecycleTarget), s.ReprovisionInstall)
		installsAccess.POST("/reprovision-stack", require.Point(permissions.InstallProvision, s.installLifecycleTarget), s.ReprovisionInstallStack)
		installsAccess.POST("/deprovision", require.Point(permissions.InstallDelete, s.installTarget), s.DeprovisionInstall)
		installsAccess.POST("/forget", require.Point(permissions.InstallDelete, s.installTarget), s.ForgetInstall)
		s.POST(installsAccess.For(require.All(s.installWorkflowRetryRequirements)), "/retry-workflow", s.RetryWorkflow, api.APIContextTypePublic, true) // Deprecated

		// install deploys
		s.GET(installsAccess.For(require.Point(permissions.InstallHistoryRead, s.installIDTarget)), "/deploys", s.GetInstallDeploys, api.APIContextTypePublic, true)           // Deprecated
		s.POST(installsAccess.For(require.Point(permissions.InstallDeploy, s.installDeployTarget)), "/deploys", s.CreateInstallDeploy, api.APIContextTypePublic, true)         // Deprecated
		s.GET(installsAccess.For(require.Point(permissions.InstallHistoryRead, s.installTarget)), "/deploys/latest", s.GetInstallLatestDeploy, api.APIContextTypePublic, true) // Deprecated
		s.GET(installsAccess.For(require.Point(permissions.InstallHistoryRead, s.installIDTarget)), "/deploys/:deploy_id", s.GetInstallDeploy, api.APIContextTypePublic, true)
		// Deprecated
		installsAccess.GET("/components/deploys", require.Point(permissions.InstallHistoryRead, s.installIDTarget), s.GetInstallComponentsDeploys)
		installsAccess.GET("/components/:component_id/deploys/:deploy_id", require.Point(permissions.InstallHistoryRead, s.installIDTarget), s.GetInstallComponentDeploy)
		// install readme
		installsAccess.GET("/readme", require.Point(permissions.InstallConfigRead, s.installTarget), s.GetInstallReadme)
		// install drifts
		installsAccess.GET("/drifted-objects", require.Point(permissions.InstallStateRead, s.installIDTarget), s.GetDriftedObjects)
		// live component resource explorer
		installsAccess.GET("/status", require.Point(permissions.InstallStateRead, s.installIDTarget), s.GetInstallStatus)
		installsAccess.GET("/resources", require.Point(permissions.InstallStateRead, s.installIDTarget), s.GetInstallResources)
		// install-level component health rollup
		installsAccess.GET("/health/timeline", require.Point(permissions.InstallStateRead, s.installIDTarget), s.GetInstallHealthTimeline)
		installsAccess.POST("/health/baseline", require.Point(permissions.InstallConfigUpdate, s.installIDTarget), s.ResetInstallHealthBaseline)
		installsAccess.POST("/health/cluster-access", require.Point(permissions.InstallConfigUpdate, s.installIDTarget), s.RefreshInstallHealthClusterAccess)
		// install state
		installsAccess.GET("/state", require.Point(permissions.InstallStateRead, s.installIDTarget), s.GetInstallState)
		installsAccess.GET("/state-history", require.Point(permissions.InstallStateRead, s.installIDTarget), s.GetInstallStateHistory)
		// install dns delegation check
		installsAccess.GET("/dns/check", require.Point(permissions.InstallStateRead, s.installTarget), s.CheckInstallDNSDelegation)
		// install sandbox
		installsAccess.POST("/reprovision-sandbox", require.Point(permissions.InstallProvision, s.installLifecycleTarget), s.ReprovisionInstallSandbox)
		installsAccess.POST("/deprovision-sandbox", require.Point(permissions.InstallTeardown, s.installLifecycleTarget), s.DeprovisionInstallSandbox)

		sandboxRuns := require.Routes(installs.Group("/sandbox-runs"), permissions.APICtl)
		{
			sandboxRuns.GET("", require.Point(permissions.InstallHistoryRead, s.installIDTarget), s.GetInstallSandboxRuns)
			sandboxRuns.GET("/:run_id", require.Point(permissions.InstallHistoryRead, s.sandboxRunInstallTarget), s.GetInstallSandboxRunV2)
		}

		// install inputs
		inputs := require.Routes(installs.Group("/inputs"), permissions.APICtl)
		{
			inputs.GET("", require.Point(permissions.InstallConfigRead, s.installIDTarget), s.GetInstallInputs)
			inputs.POST("", require.Point(permissions.InstallInputsUpdate, s.installLifecycleTarget), s.CreateInstallInputs)
			inputs.GET("/current", require.Point(permissions.InstallConfigRead, s.installIDTarget), s.GetInstallCurrentInputs)
			inputs.PATCH("", require.Point(permissions.InstallInputsUpdate, s.installLifecycleTarget), s.UpdateInstallInputs)
		}

		// install components
		components := installs.Group("/components")
		componentsAccess := require.Routes(components, permissions.APICtl)
		{
			componentsAccess.GET("", require.Point(permissions.InstallInspect, s.installIDTarget), s.GetInstallComponents)
			componentsAccess.POST("/teardown-all", require.Point(permissions.InstallTeardown, s.installLifecycleTarget), s.TeardownInstallComponents)
			componentsAccess.POST("/deploy-all", require.Point(permissions.InstallDeploy, s.installLifecycleTarget), s.DeployInstallComponents)

			component := require.Routes(components.Group("/:component_id"), permissions.APICtl)
			{
				component.GET("", require.Point(permissions.InstallInspect, s.installIDTarget), s.GetInstallComponent)
				component.POST("/teardown", require.Point(permissions.InstallTeardown, s.installComponentTarget), s.TeardownInstallComponent)
				component.POST("/toggle", require.Point(permissions.InstallConfigUpdate, s.installComponentTarget), s.ToggleInstallComponent)
				component.POST("/forget", require.Point(permissions.InstallTeardown, s.installTarget), s.ForgetInstallComponent)
				component.GET("/deploys", require.Point(permissions.InstallHistoryRead, s.installIDTarget), s.GetInstallComponentDeploys)
				component.GET("/outputs", require.Point(permissions.InstallStateRead, s.installIDTarget), s.GetInstallComponentOutputs)
				component.GET("/deploys/latest", require.Point(permissions.InstallHistoryRead, s.installIDTarget), s.GetInstallComponentLatestDeploy)
				component.POST("/deploys", require.Point(permissions.InstallDeploy, s.installDeployTarget), s.CreateInstallComponentDeploy)
				component.POST("/recover-helm-release", require.Point(permissions.InstallDeploy, s.installComponentTarget), s.RecoverInstallComponentHelmRelease)
				// component health: gin can't mix wildcard names at the same path
				// depth, so these reuse the ":component_id" node above, but the
				// value they expect is the install_component's own ID (matching
				// the ClickHouse rows), not the catalog component ID the sibling
				// routes above use. The Swagger docs name it install_component_id.
				component.GET("/health/timeline", require.Point(permissions.InstallStateRead, s.installIDTarget), s.GetInstallComponentHealthTimeline)
				component.GET("/health/incident", require.Point(permissions.InstallStateRead, s.installIDTarget), s.GetInstallComponentHealthIncident)
				component.GET("/health/checks", require.Point(permissions.InstallStateRead, s.installIDTarget), s.GetInstallComponentHealthChecks)
				component.PUT("/health/checks/:check_name", require.Point(permissions.InstallConfigUpdate, s.installIDTarget), s.PutInstallComponentHealthCheck)
			}
		}

		// install action workflows
		actions := installs.Group("/actions")
		{
			action := require.Routes(actions.Group("/:action_id"), permissions.APICtl)
			{
				action.GET("/outputs", require.Point(permissions.InstallStateRead, s.installIDTarget), s.GetInstallActionWorkflowOutputs)
			}
		}
		installsAccess.POST("/sync-secrets", require.Point(permissions.InstallSecretsSync, s.installLifecycleTarget), s.SyncSecrets)
		installsAccess.POST("/sync-config", require.Point(permissions.InstallConfigUpdate, s.installIDTarget), s.SyncInstallConfig)

		// install events
		events := installsAccess.For(require.Point(permissions.InstallHistoryRead, s.installIDTarget)).Group("/events")
		{
			events.GET("", s.GetInstallEvents)
			events.GET("/:event_id", s.GetInstallEvent)
		}
		// workflows for install
		installsAccess.GET("/workflows", require.Collection(permissions.WorkflowInspect, s.workflowResources), s.GetWorkflows)
		// install runner group
		installsAccess.GET("/runner-group", require.Point(permissions.InstallInspect, s.installIDTarget), s.GetInstallRunnerGroup)
		installsAccess.GET("/telemetry", require.Point(permissions.InstallConfigRead, s.installIDTarget), s.GetInstallTelemetrySettings)
		installsAccess.POST("/telemetry/collector", require.Point(permissions.InstallCollectorManage, s.installIDTarget), s.EnsureInstallTelemetryCollector)
		installsAccess.GET("/telemetry/collector", require.Point(permissions.InstallInspect, s.installIDTarget), s.GetInstallTelemetryCollector)
		installsAccess.DELETE("/telemetry/collector", require.Point(permissions.InstallCollectorManage, s.installIDTarget), s.DeleteInstallTelemetryCollector)
		// runner bootstrap token
		installsAccess.POST("/runner-bootstrap-token", require.Point(permissions.InstallRunnerBootstrap, s.installTarget), s.CreateRunnerBootstrapToken)
		// install stacks
		installsAccess.GET("/stack", require.Point(permissions.InstallConfigRead, s.installIDTarget), s.GetInstallStackByInstallID)
		installsAccess.GET("/stack-runs", require.Point(permissions.InstallHistoryRead, s.installIDTarget), s.GetInstallStackRuns)
		// available roles
		installsAccess.GET("/available-roles", require.Point(permissions.InstallConfigRead, s.installLifecycleTarget), s.GetAvailableRoles)
		// app permissions config with provisioning status
		installsAccess.GET("/app-permissions-config", require.Point(permissions.InstallConfigRead, s.installLifecycleTarget), s.GetInstallAppPermissionsConfig)

		// install roles
		roles := require.Routes(installs.Group("/roles"), permissions.APICtl)
		{
			roles.GET("", require.Point(permissions.InstallConfigRead, s.installIDTarget), s.GetInstallRoles)
			roles.GET("/latest", require.Point(permissions.InstallConfigRead, s.installIDTarget), s.GetLatestInstallRoles)
			roles.GET("/usages", require.Point(permissions.InstallHistoryRead, s.installIDTarget), s.GetInstallRoleUsages)
			roles.PATCH("/:role_id", require.Point(permissions.InstallConfigUpdate, s.installIDTarget), s.UpdateInstallRole)
		}

		// install config
		configs := installsAccess.For(require.Point(permissions.InstallConfigUpdate, s.installIDTarget)).Group("/configs")
		{
			configs.POST("", s.CreateInstallConfig)
			configs.PATCH("/:config_id", s.UpdateInstallConfig)
		}
		// install app config versions
		installsAccess.GET("/app-config-versions", require.Point(permissions.InstallConfigRead, s.installIDTarget), s.GetInstallAppConfigVersions)
		installsAccess.GET("/app-config-versions/:version_id/diff", require.Point(permissions.InstallConfigRead, s.installIDTarget), s.GetInstallAppConfigVersionDiff)
		installsAccess.GET("/app-configs/:config_id/diff", require.Point(permissions.InstallConfigRead, s.installIDTarget), s.GetInstallAppConfigTreeDiff)
		installsAccess.GET("/deployments", require.Point(permissions.InstallHistoryRead, s.installIDTarget), s.GetInstallDeployments)
		installsAccess.GET("/deployments/:workflow_id", require.Point(permissions.InstallHistoryRead, s.installIDTarget), s.GetInstallDeployment)
		installsAccess.GET("/deployments/:workflow_id/source-diff", require.Point(permissions.InstallHistoryRead, s.installIDTarget), s.GetInstallDeploymentSourceDiff)
		installsAccess.GET("/deployment-summaries", require.Point(permissions.InstallHistoryRead, s.installIDTarget), s.GetInstallDeploymentSummaries)
		installsAccess.GET("/overview", require.Point(permissions.InstallStateRead, s.installIDTarget), s.GetInstallOverview)
		installsAccess.POST("/resolve-drift", require.Point(permissions.InstallDriftResolve, s.installIDTarget), s.ResolveInstallDrift)
		installsAccess.GET("/activity", require.Point(permissions.InstallHistoryRead, s.installIDTarget), s.GetInstallActivity)
		installsAccess.GET("/updates", require.Point(permissions.InstallConfigRead, s.installIDTarget), s.GetInstallUpdates)
		installsAccess.GET("/config-versions", require.Point(permissions.InstallConfigRead, s.installIDTarget), s.GetInstallConfigVersions)
		installsAccess.GET("/config-versions/:version_id/diff", require.Point(permissions.InstallConfigRead, s.installIDTarget), s.GetInstallConfigVersionDiff)
		installsAccess.GET("/config-syncs", require.Point(permissions.InstallConfigRead, s.installIDTarget), s.GetInstallConfigSyncs)
		installsAccess.POST("/app-config-updates", require.Point(permissions.InstallConfigUpdate, s.installLifecycleTarget), s.CreateInstallAppConfigUpdate)
		// install audit logs
		installsAccess.GET("/audit_logs", require.Point(permissions.InstallHistoryRead, s.installIDTarget), s.GetInstallAuditLogs)
		// install cli config
		installsAccess.GET("/generate-cli-install-config", require.Point(permissions.InstallConfigRead, s.installLifecycleTarget), s.GenerateCLIInstallConfig)
	}
	// stack lookup by stack_id
	orgAccess.GET("/v1/installs/stacks/:stack_id", require.Point(permissions.InstallConfigRead, s.stackInstallTarget), s.GetInstallStackByStackID)
	// org-level workflow queries (must be registered before /:workflow_id group)
	orgAccess.GET("/v1/workflows/pending-approvals", require.Collection(permissions.WorkflowInspect, s.workflowResources), s.GetOrgPendingApprovals)
	orgAccess.GET("/v1/workflows", require.Collection(permissions.WorkflowInspect, s.workflowResources), s.GetOrgWorkflows)
	orgAccess.POST("/v1/workflows/cancel", require.Collection(permissions.WorkflowCancel, s.workflowResources), s.CancelWorkflows)

	// workflows (standalone)
	workflows := require.OrgRoute(ge).Group("/v1/workflows/:workflow_id", require.WorkflowInOrg(s.db))
	workflowsAccess := require.Routes(workflows, permissions.APICtl)
	{
		workflowsAccess.GET("", require.Point(permissions.WorkflowInspect, s.workflowTarget), s.GetWorkflow)
		workflowsAccess.PATCH("", require.All(s.workflowContinuation(permissions.WorkflowApprove)), s.UpdateWorkflow)
		workflowsAccess.POST("/cancel", require.Point(permissions.WorkflowCancel, s.workflowTarget), s.CancelWorkflow)
		workflowsAccess.GET("/queue-position", require.Point(permissions.WorkflowInspect, s.workflowTarget), s.GetWorkflowQueuePosition)

		stepGroups := workflowsAccess.For(require.Point(permissions.WorkflowInspect, s.workflowTarget)).Group("/step-groups")
		{
			stepGroups.GET("", s.GetWorkflowStepGroups)
			stepGroups.GET("/:step_group_id", s.GetWorkflowStepGroup)
		}

		steps := workflows.Group("/steps")
		stepsAccess := require.Routes(steps, permissions.APICtl)
		{
			stepsAccess.GET("", require.Point(permissions.WorkflowInspect, s.workflowTarget), s.GetWorkflowSteps)
			stepsAccess.GET("/:step_id", require.Point(permissions.WorkflowInspect, s.workflowTarget), s.GetWorkflowStep)
			stepsAccess.GET("/:step_id/await", require.Point(permissions.WorkflowInspect, s.workflowTarget), s.AwaitWorkflowStep)
			stepsAccess.POST("/:step_id/retry", require.All(s.workflowContinuation(permissions.WorkflowRetry)), s.RetryWorkflowStep)
			stepsAccess.POST("/:step_id/skip", require.All(s.workflowContinuation(permissions.WorkflowRetry)), s.SkipWorkflowStep)
			stepsAccess.POST("/:step_id/cancel", require.Point(permissions.WorkflowCancel, s.workflowTarget), s.CancelWorkflowStep)

			approvals := require.Routes(steps.Group("/:step_id/approvals/:approval_id"), permissions.APICtl)
			{
				approvals.GET("", require.Point(permissions.WorkflowInspect, s.workflowTarget), s.GetWorkflowStepApproval)
				approvals.POST("/response", require.All(s.workflowApprovalRequirements), s.CreateWorkflowStepApprovalResponse)
				approvals.GET("/contents", require.Point(permissions.WorkflowInspect, s.workflowTarget), s.GetWorkflowStepApprovalContents)
			}
		}
	}

	// deprecated install-workflows
	installWorkflows := require.Routes(org.Group("/v1/install-workflows/:install_workflow_id", require.InstallWorkflowInOrg(s.db)), permissions.APICtl)
	s.GET(installWorkflows.For(require.Point(permissions.WorkflowInspect, s.workflowTarget)), "", s.GetInstallWorkflow, api.APIContextTypePublic, true)
	s.PATCH(installWorkflows.For(require.All(s.workflowContinuation(permissions.WorkflowApprove))), "", s.UpdateInstallWorkflow, api.APIContextTypePublic, true)
	s.GET(installWorkflows.For(require.Point(permissions.WorkflowInspect, s.workflowTarget)), "/steps", s.GetInstallWorkflowSteps, api.APIContextTypePublic, true)
	s.GET(installWorkflows.For(require.Point(permissions.WorkflowInspect, s.workflowTarget)), "/steps/:install_workflow_step_id", s.GetInstallWorkflowStep, api.APIContextTypePublic, true)
	s.POST(installWorkflows.For(require.Point(permissions.WorkflowCancel, s.workflowTarget)), "/cancel", s.CancelInstallWorkflow, api.APIContextTypePublic, true)
	s.GET(installWorkflows.For(require.Point(permissions.WorkflowInspect, s.workflowTarget)), "/steps/:install_workflow_step_id/approvals/:approval_id", s.GetInstallWorkflowStepApproval, api.APIContextTypePublic, true)

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
			install.POST("/admin-cleanup-tokens", s.AdminCleanupInstallTokens)

			// NOTE(JM): the following endpoints should be removed after workflows/independent runners are rolled out
			install.POST("/admin-reprovision", s.ReprovisionInstall)
			install.POST("/admin-forget", s.AdminForgetInstall)
			install.POST("/admin-unforget", s.AdminUnforgetInstall)
			install.POST("/admin-update-sandbox", s.AdminUpdateSandbox)
		}
	}

	// orgs
	orgs := api.Group("/v1/orgs/:org_id")
	{
		orgs.POST("/admin-forget-installs", s.ForgetOrgInstalls)
		orgs.GET("/admin-get-installs", s.AdminGetOrgInstalls)
		orgs.GET("/admin-get-forgotten-installs", s.AdminGetForgottenInstalls)
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

func (s *service) RegisterRunnerRoutes(engine *gin.Engine) error {
	s.POST(engine, "/v1/installs/:install_id/phone-home/:phone_home_id", s.InstallPhoneHomeRunner, api.APIContextTypeRunner, true)
	settings := require.Routes(require.Route(engine, permissions.KindTelemetry, permissions.PermissionRead, "install_id"), permissions.APIRunner)
	settings.GET("/v1/installs/:install_id/telemetry/collector-settings", require.Point(permissions.TelemetrySettingsRead, s.installIDTarget), s.GetInstallTelemetryCollectorSettings)
	tokens := require.Routes(require.Route(engine, permissions.KindTelemetry, permissions.PermissionCreate, "install_id"), permissions.APIRunner)
	tokens.POST("/v1/installs/:install_id/telemetry/access-token", require.Point(permissions.TelemetryTokenCreate, s.installIDTarget), s.CreateInstallTelemetryAccessToken)
	return nil
}

func (s *service) RegisterAuthRoutes(api *gin.Engine) error {
	return nil
}

func (s *service) RegisterAdminDashboardRoutes(api *gin.Engine) error {
	return nil
}

func New(params Params) (*service, error) {
	issuer, err := telemetrytoken.New(params.Cfg)
	if err != nil {
		return nil, fmt.Errorf("initialize telemetry token issuer: %w", err)
	}

	return &service{
		RouteRegister: api.RouteRegister{
			EndpointAudit: params.EndpointAudit,
		},
		cfg:                  params.Cfg,
		l:                    params.L,
		v:                    params.V,
		db:                   params.DB,
		chDB:                 params.CHDB,
		mw:                   params.MW,
		componentHelpers:     params.ComponentHelpers,
		helpers:              params.Helpers,
		alignment:            alignment.New(params.DB, params.Helpers, params.BlobService, params.L),
		accountsHelpers:      params.AccountsHelpers,
		acctClient:           params.AcctClient,
		queueClient:          params.QueueClient,
		appsHelpers:          params.AppsHelpers,
		orgsHelpers:          params.OrgsHelpers,
		runnersHelpers:       params.RunnersHelpers,
		actionsHelpers:       params.ActionsHelpers,
		featuresClient:       params.FeaturesClient,
		flowsClient:          params.FlowsClient,
		blobSvc:              params.BlobService,
		audit:                params.Audit,
		workloadJWT:          workloadjwt.NewVerifier(),
		telemetryTokenIssuer: issuer,
	}, nil
}

func (s *service) RegisterSlackRoutes(api *gin.Engine) error {
	return nil
}
