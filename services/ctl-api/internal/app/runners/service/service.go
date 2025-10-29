package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/account"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/api"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
)

type Params struct {
	fx.In

	V             *validator.Validate
	Cfg           *internal.Config
	DB            *gorm.DB `name:"psql"`
	CHDB          *gorm.DB `name:"ch"`
	MW            metrics.Writer
	L             *zap.Logger
	EvClient      eventloop.Client
	AccountClient *account.Client
	Helpers       *helpers.Helpers
}

type service struct {
	v          *validator.Validate
	l          *zap.Logger
	db         *gorm.DB
	chDB       *gorm.DB
	mw         metrics.Writer
	cfg        *internal.Config
	evClient   eventloop.Client
	acctClient *account.Client
	helpers    *helpers.Helpers
}

var _ api.Service = (*service)(nil)

func (s *service) RegisterPublicRoutes(api *gin.Engine) error {
	api.GET("/v1/runners/:runner_id", s.GetRunnerCtlAPI)
	api.GET("/v1/runners/:runner_id/connected", s.GetRunnerConnectStatus)
	api.GET("/v1/runners/:runner_id/jobs", s.GetRunnerJobsCtlAPI)
	api.GET("/v1/runner-jobs/:runner_job_id/plan", s.GetRunnerJobPlan)
	api.GET("/v1/runner-jobs/:runner_job_id/composite-plan", s.GetRunnerJobCompositePlan)
	api.POST("/v1/runner-jobs/:runner_job_id/cancel", s.CancelRunnerJob)
	api.GET("/v1/runner-jobs/:runner_job_id", s.GetRunnerJob)
	api.GET("/v1/runners/:runner_id/recent-health-checks", s.GetRunnerRecentHealthChecks)
	api.GET("/v1/runners/:runner_id/latest-heart-beat", s.GetRunnerLatestHeartBeat)
	api.GET("/v1/runners/:runner_id/heart-beats/latest", s.GetLatestRunnerHeartBeatFromView)

	// trigger specific jobs
	api.POST("/v1/runners/:runner_id/graceful-shutdown", s.GracefulShutDown)
	api.POST("/v1/runners/:runner_id/force-shutdown", s.ForceShutDown)
	api.POST("/v1/runners/:runner_id/mng/shutdown-vm", s.MngVMShutDown)
	api.POST("/v1/runners/:runner_id/mng/shutdown", s.MngShutDown)
	api.POST("/v1/runners/:runner_id/mng/update", s.MngUpdate)

	// settings
	api.GET("/v1/runners/:runner_id/settings", s.GetRunnerSettingsPublic)
	api.PATCH("/v1/runners/:runner_id/settings", s.UpdateRunnerSettings)

	tfWorkspacePath := "/v1/terraform-workspaces"
	api.GET(tfWorkspacePath, s.GetTerraformWorkpaces)
	api.GET(tfWorkspacePath+"/:workspace_id", s.GetTerraformWorkpace)
	api.POST(tfWorkspacePath, s.CreateTerraformWorkspaceV2)
	api.DELETE(tfWorkspacePath+"/:workspace_id", s.DeleteTerraformWorkpace)
	api.GET(tfWorkspacePath+"/:workspace_id/lock", s.GetTerraformWorkspaceLock)
	api.POST(tfWorkspacePath+"/:workspace_id/lock", s.LockTerraformWorkspace)
	api.POST(tfWorkspacePath+"/:workspace_id/unlock", s.UnlockTerraformWorkspace)
	api.GET(tfWorkspacePath+"/:workspace_id/states", s.GetTerraformWorkspaceStatesV2)
	api.GET(tfWorkspacePath+"/:workspace_id/states/:state_id", s.GetTerraformWorkspaceStateByIDV2)
	api.GET(tfWorkspacePath+"/:workspace_id/states/:state_id/resources", s.GetTerraformWorkspaceStateResourcesV2)

	api.GET(tfWorkspacePath+"/:workspace_id/state-json", s.GetTerraformWorkspaceStatesJSONV2)
	api.GET(tfWorkspacePath+"/:workspace_id/state-json/:state_id", s.GetTerraformWorkspaceStatesJSONByIDV2)
	api.GET(tfWorkspacePath+"/:workspace_id/state-json/:state_id/resources", s.GetTerraformWorkspaceStateResourcesV2)

	api.POST("/v1/terraform-workspace", s.CreateTerraformWorkspace)                                                            // deprecated
	api.GET("/v1/runners/terraform-workspace/:workspace_id/states", s.GetTerraformWorkspaceStates)                             // deprecated
	api.GET("/v1/runners/terraform-workspace/:workspace_id/states/:state_id", s.GetTerraformWorkspaceStateByID)                // deprecated
	api.GET("/v1/runners/terraform-workspace/:workspace_id/states/:state_id/resources", s.GetTerraformWorkspaceStateResources) // deprecated

	api.GET("/v1/runners/terraform-workspace/:workspace_id/state-json", s.GetTerraformWorkspaceStatesJSON)                         // deprecated
	api.GET("/v1/runners/terraform-workspace/:workspace_id/state-json/:state_id", s.GetTerraformWorkspaceStatesJSONByID)           // deprecated
	api.GET("/v1/runners/terraform-workspace/:workspace_id/state-json/:state_id/resources", s.GetTerraformWorkspaceStateResources) // deprecated

	tfBackendPath := "/v1/terraform-backend"
	api.GET(tfBackendPath, s.GetTerraformCurrentStateData)
	api.POST(tfBackendPath, s.UpdateTerraformState)

	api.GET("/v1/log-streams/:log_stream_id/logs", s.LogStreamReadLogs)
	api.GET("/v1/log-streams/:log_stream_id", s.GetLogStream)

	return nil
}
func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	// runners
	runners := api.Group("/v1/runners")
	{
		runners.GET("", s.AdminGetAllRunners)
		runners.POST("/restart", s.AdminRestartRunners)
		runners.PATCH("/bulk-update", s.AdminBulkUpdateRunners)

		// runner-specific operations
		runner := runners.Group("/:runner_id")
		{
			runner.GET("", s.AdminGetRunner)

			// runner settings
			runner.GET("/settings", s.AdminGetRunnerSettings)
			runner.PATCH("/settings", s.AdminUpdateRunnerSettings)

			// runner lifecycle
			runner.POST("/reprovision", s.AdminReprovisionRunner)
			runner.POST("/deprovision", s.AdminDeprovisionRunner)
			runner.POST("/delete", s.AdminDeleteRunner)
			runner.POST("/force-delete", s.AdminForceDeleteRunner)
			runner.POST("/restart", s.RestartRunner)
			runner.POST("/offline-check", s.AdminOfflineCheck)

			// service account management
			runner.POST("/service-account-token", s.AdminCreateRunnerServiceAccountToken)
			runner.POST("/invalidate-service-account-token", s.AdminInvalidateRunnerServiceAccountToken)
			runner.POST("/extend-service-account-token", s.AdminExtendRunnerServiceAccountToken)
			runner.GET("/service-account", s.AdminGetRunnerServiceAccount)

			// job management
			runner.POST("/flush-orphaned-jobs", s.AdminFlushOrphanedJobs)
			runner.GET("/jobs/queue", s.AdminGetRunnerJobsQueue)

			// trigger specific jobs
			runner.POST("/graceful-shutdown", s.AdminGracefulShutDown)
			runner.POST("/force-shutdown", s.AdminForceShutDown)
			runner.POST("/noop-job", s.AdminCreateNoopJob)
			runner.POST("/health-check-job", s.AdminCreateHealthCheck)
		}
	}

	// runner groups
	runnerGroups := api.Group("/v1/runner-groups/:runner_group_id")
	{
		runnerGroups.GET("", s.AdminGetRunnerGroup)
	}

	// runner job management
	runnerJobs := api.Group("/v1/runner-jobs/:runner_job_id")
	{
		runnerJobs.POST("/cancel", s.AdminCancelRunnerJob)
		runnerJobs.GET("", s.AdminGetRunnerJob)
	}

	// otel admin endpoints
	logStreams := api.Group("/v1/log-streams/:log_stream_id")
	{
		logStreams.GET("/logs", s.AdminGetLogStreamLogs)
		logStreams.GET("", s.AdminGetLogStream)
	}

	// install runners
	installs := api.Group("/v1/installs/:install_id")
	{
		installs.POST("/runners/shutdown-job", s.AdminCreateInstallRunnerqShutDownJob)
	}

	// terraform workspace management
	workspaces := api.Group("/v1/terraform-workspaces/:workspace_id")
	{
		workspaces.POST("/lock", s.AdminLockWorkspace)
		workspaces.POST("/unlock", s.AdminUnlockWorkspace)
	}

	return nil
}

func (s *service) RegisterRunnerRoutes(api *gin.Engine) error {
	runners := api.Group("/v1/runners/:runner_id")
	runners.POST("/health-checks", s.CreateRunnerHealthCheck)
	runners.POST("/heart-beats", s.CreateRunnerHeartBeat)
	runners.GET("", s.GetRunner)
	runners.GET("/jobs", s.GetRunnerJobs)
	runners.GET("/settings", s.GetRunnerSettings)
	runners.POST("/traces", s.OtelWriteTraces)
	runners.POST("/metrics", s.OtelWriteMetrics)
	runners.GET("/jobs/:job_id/plan", s.GetRunnerJobPlanV2)
	runners.GET("/jobs/:job_id", s.GetRunnerJobV2)
	runners.PATCH("", s.UpdateRunnerJobV2)

	runnerJobs := api.Group("/v1/runner-jobs/:runner_job_id")
	runnerJobs.GET("", s.GetRunnerJob)          //deprecated
	runnerJobs.PATCH("", s.UpdateRunnerJob)     //deprecated
	runnerJobs.GET("/plan", s.GetRunnerJobPlan) //deprecated

	executions := runnerJobs.Group("/executions")
	executions.POST("", s.CreateRunnerJobExecution)
	executions.GET("", s.GetRunnerJobExecutions)
	executions.GET("/:runner_job_execution_id", s.GetRunnerJobExecution)
	executions.PATCH("/:runner_job_execution_id", s.UpdateRunnerJobExecution)
	executions.POST("/:runner_job_execution_id/result", s.CreateRunnerJobExecutionResult)
	executions.POST("/:runner_job_execution_id/outputs", s.CreateRunnerJobExecutionOutputs)

	// Terraform backend
	tfBackend := api.Group("/v1/terraform-backend")
	tfBackend.GET("", s.GetTerraformCurrentStateData)
	tfBackend.POST("", s.UpdateTerraformState)
	tfBackend.DELETE("", s.DeleteTerraformState)

	// terraform workspaces
	tfWorkspaces := api.Group("/v1/terraform-workspaces")
	tfWorkspaces.GET("", s.GetTerraformWorkpaces)
	tfWorkspaces.POST("", s.CreateTerraformWorkspace)
	tfWorkspaces.GET("/:workspace_id", s.GetTerraformWorkpace)
	tfWorkspaces.DELETE("/:workspace_id", s.DeleteTerraformWorkpace)
	tfWorkspaces.POST("/:workspace_id/lock", s.LockTerraformWorkspace)
	tfWorkspaces.POST("/:workspace_id/unlock", s.UnlockTerraformWorkspace)
	// terraform state json
	tfWorkspaces.POST("/:workspace_id/state-json", s.UpdateTerraformWorkspaceStateJSON)
	tfWorkspaces.DELETE("/:workspace_id/states", s.DeleteTerraformWorkspaceStateJSON)

	// helm release api
	helmReleasePath := "/v1/helm-releases/:helm_chart_id/releases/"
	api.GET(helmReleasePath+":namespace", s.GetHelmReleases)
	api.GET(helmReleasePath+":namespace/:key", s.GetHelmRelease)
	api.GET(helmReleasePath+":namespace/query", s.QueryHelmRelease)
	api.POST(helmReleasePath+":namespace/:key", s.CreateHelmRelease)
	api.PUT(helmReleasePath+":namespace/:key", s.UpdateHelmRelease)
	api.DELETE(helmReleasePath+":namespace/:key", s.DeleteHelmRelease)

	// TODO(jm): these will be moved to the otel namespace
	api.POST("/v1/log-streams/:log_stream_id/logs", s.LogStreamWriteLogs)

	// installs
	installs := api.Group("/v1/installs")
	installs.GET("/:install_id/:component_id/last-active-plan", s.GetInstallComponenetLastActivePlan)

	return nil
}

func New(params Params) *service {
	return &service{
		cfg:        params.Cfg,
		l:          params.L,
		v:          params.V,
		db:         params.DB,
		chDB:       params.CHDB,
		mw:         params.MW,
		evClient:   params.EvClient,
		acctClient: params.AccountClient,
		helpers:    params.Helpers,
	}
}
