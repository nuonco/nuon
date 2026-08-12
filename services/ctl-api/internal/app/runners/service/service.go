package service

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/pkg/errors"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/metrics"
	temporalclient "github.com/nuonco/nuon/pkg/temporal/client"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/heartbeater"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/kafka"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	emitterclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/client"
)

type Params struct {
	fx.In

	V                    *validator.Validate
	Cfg                  *internal.Config
	DB                   *gorm.DB `name:"psql"`
	CHDB                 *gorm.DB `name:"ch"`
	MW                   metrics.Writer
	L                    *zap.Logger
	AccountClient        *account.Client
	Helpers              *helpers.Helpers
	InstallsHelpers      *installshelpers.Helpers
	EndpointAudit        *apiPkg.EndpointAudit
	RunnerHeartbeatCache *RunnerHeartbeatCache
	Heartbeater          *heartbeater.Heartbeater
	Kafka                *kafka.Producer
	FeaturesClient       *features.Features
	TemporalClient       temporalclient.Client
	RunnerJobWake        *RunnerJobWakeRegistry
	BlobSvc              blobstore.Service
	EmitterClient        *emitterclient.Client
	QueueClient          *queueclient.Client
}

type service struct {
	apiPkg.RouteRegister
	v                    *validator.Validate
	l                    *zap.Logger
	db                   *gorm.DB
	chDB                 *gorm.DB
	mw                   metrics.Writer
	cfg                  *internal.Config
	acctClient           *account.Client
	helpers              *helpers.Helpers
	installsHelpers      *installshelpers.Helpers
	runnerHeartbeatCache *RunnerHeartbeatCache
	heartbeater          *heartbeater.Heartbeater
	kafka                *kafka.Producer
	featuresClient       *features.Features
	temporalClient       temporalclient.Client
	runnerJobWake        *RunnerJobWakeRegistry
	blobSvc              blobstore.Service
	emitterClient        *emitterclient.Client
	queueClient          *queueclient.Client
	// logStreamCache hits in front of getLogStream on the OTLP ingest
	// hot path. The fields the writer reads (OwnerType, ParentLogStreamID)
	// are effectively immutable for the life of the stream, so a 5min TTL
	// trades a 5min staleness window for one fewer Postgres round-trip
	// per OTLP batch.
	logStreamCache *expirable.LRU[string, *app.LogStream]

	// ensuredHealthQueues memoizes which installs this process has reconciled
	// queues for from the component-health ingest path, so installs that
	// predate the health evaluator's queue get it lazily instead of requiring
	// an admin queue migration.
	ensuredHealthQueues sync.Map
}

const (
	logStreamCacheSize = 10_000
	logStreamCacheTTL  = 5 * time.Minute
)

var _ apiPkg.Service = (*service)(nil)

func (s *service) RegisterPublicRoutes(api *gin.Engine) error {
	api.GET("/v1/runners/:runner_id", s.GetRunnerCtlAPI)
	api.GET("/v1/runners/:runner_id/connected", s.GetRunnerConnectStatus)
	api.GET("/v1/runners/:runner_id/jobs", s.GetRunnerJobsCtlAPI)
	api.GET("/v1/runner-jobs", s.ListRunnerJobsCtlAPI)
	api.GET("/v1/runner-jobs/:runner_job_id/plan", s.GetRunnerJobPlanPublic)
	api.GET("/v1/runner-jobs/:runner_job_id/composite-plan", s.GetRunnerJobCompositePlan)
	api.POST("/v1/runner-jobs/:runner_job_id/cancel", s.CancelRunnerJob)
	api.GET("/v1/runner-jobs/:runner_job_id", s.GetRunnerJobPublic)
	api.GET("/v1/runners/:runner_id/recent-health-checks", s.GetRunnerRecentHealthChecks)
	api.GET("/v1/runners/:runner_id/latest-heart-beat", s.GetRunnerLatestHeartBeat)
	api.GET("/v1/runners/:runner_id/heart-beats/latest", s.GetLatestRunnerHeartBeatFromView)

	api.GET("/v1/runners/:runner_id/card-details", s.GetRunnerCardDetails)

	// runner process endpoints
	api.GET("/v1/runners/:runner_id/processes", s.ListRunnerProcesses)
	api.GET("/v1/runners/:runner_id/processes/current", s.GetCurrentRunnerProcesses)
	api.GET("/v1/runners/:runner_id/processes/:process_id", s.GetRunnerProcessPublic)
	api.GET("/v1/runners/:runner_id/processes/:process_id/heart-beats/latest", s.GetProcessLatestHeartBeat)
	api.POST("/v1/runners/:runner_id/processes/:process_id/shutdown", s.ShutdownRunnerProcess)

	// trigger specific jobs
	api.POST("/v1/runners/:runner_id/graceful-shutdown", s.GracefulShutDown)
	api.POST("/v1/runners/:runner_id/force-shutdown", s.ForceShutDown)
	api.POST("/v1/runners/:runner_id/mng/shutdown-vm", s.MngVMShutDown)
	api.POST("/v1/runners/:runner_id/mng/shutdown", s.MngShutDown)
	api.POST("/v1/runners/:runner_id/mng/update", s.MngUpdate)
	api.POST("/v1/runners/:runner_id/mng/restart", s.MngRestart)
	api.POST("/v1/runners/:runner_id/prune-tokens", s.PruneTokens)

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
	api.GET(tfWorkspacePath+"/:workspace_id/state-json/:state_id/raw", s.GetWorkspaceStateJSONRawByID)
	api.GET(tfWorkspacePath+"/:workspace_id/state-json/:state_id/resources", s.GetTerraformWorkspaceStateResourcesV2)

	s.POST(api, "/v1/terraform-workspace", s.CreateTerraformWorkspace, apiPkg.APIContextTypePublic, true)
	s.GET(api, "/v1/runners/terraform-workspace/:workspace_id/states", s.GetTerraformWorkspaceStates, apiPkg.APIContextTypePublic, true)
	s.GET(api, "/v1/runners/terraform-workspace/:workspace_id/states/:state_id", s.GetTerraformWorkspaceStateByID, apiPkg.APIContextTypePublic, true)
	s.GET(api, "/v1/runners/terraform-workspace/:workspace_id/states/:state_id/resources", s.GetTerraformWorkspaceStateResources, apiPkg.APIContextTypePublic, true)

	s.GET(api, "/v1/runners/terraform-workspace/:workspace_id/state-json", s.GetTerraformWorkspaceStatesJSON, apiPkg.APIContextTypePublic, true)
	s.GET(api, "/v1/runners/terraform-workspace/:workspace_id/state-json/:state_id", s.GetTerraformWorkspaceStatesJSONByID, apiPkg.APIContextTypePublic, true)
	s.GET(api, "/v1/runners/terraform-workspace/:workspace_id/state-json/:state_id/raw", s.GetWorkspaceStateJSONRawByID, apiPkg.APIContextTypePublic, true)
	s.GET(api, "/v1/runners/terraform-workspace/:workspace_id/state-json/:state_id/resources", s.GetTerraformWorkspaceStateResources, apiPkg.APIContextTypePublic, true)

	tfBackendPath := "/v1/terraform-backend"
	api.GET(tfBackendPath, s.GetTerraformCurrentStateData)
	api.POST(tfBackendPath, s.UpdateTerraformState)

	api.GET("/v1/log-streams/:log_stream_id/logs", s.LogStreamReadLogs)
	api.GET("/v1/log-streams/:log_stream_id/logs/tail", s.LogStreamTailLogs)
	api.GET("/v1/log-streams/:log_stream_id/spans", s.LogStreamReadSpans)
	api.GET("/v1/log-streams/:log_stream_id", s.GetLogStream)

	// install-nested, ancestor-scoped access to a runner job's logs and to
	// terraform workspaces owned by the install (bare routes above stay org-tier).
	jobs := api.Group("/v1/installs/:install_id/runner-jobs/:runner_job_id")
	jobs.Use(s.requireRunnerJobInInstall)

	// Annotated wrappers (see nested.go) so each path is a distinct swagger
	// operation and gets a generated client method; the group guard is what
	// makes them ancestor-scoped.
	jobs.GET("", s.GetInstallRunnerJob)
	jobs.GET("/plan", s.GetInstallRunnerJobPlan)
	jobs.GET("/composite-plan", s.GetInstallRunnerJobCompositePlan)
	jobs.POST("/cancel", s.CancelInstallRunnerJob)

	jobs.GET("/logs", s.ReadInstallRunnerJobLogs)
	jobs.GET("/logs/tail", s.TailInstallRunnerJobLogs)
	jobs.GET("/spans", s.ReadInstallRunnerJobSpans)

	ws := api.Group("/v1/installs/:install_id/terraform-workspaces/:workspace_id")
	ws.Use(s.requireTerraformWorkspaceInInstall)

	ws.GET("", s.GetInstallTerraformWorkspace)
	ws.GET("/lock", s.GetInstallTerraformWorkspaceLock)
	ws.POST("/lock", s.LockInstallTerraformWorkspace)
	ws.POST("/unlock", s.UnlockInstallTerraformWorkspace)

	ws.GET("/states", s.GetInstallTerraformWorkspaceStates)
	ws.GET("/states/:state_id", s.GetInstallTerraformWorkspaceState)
	ws.GET("/states/:state_id/resources", s.GetInstallTerraformWorkspaceStateResources)

	ws.GET("/state-json", s.GetInstallTerraformWorkspaceStatesJSON)
	ws.GET("/state-json/:state_id", s.GetInstallTerraformWorkspaceStateJSON)
	ws.GET("/state-json/:state_id/raw", s.GetInstallTerraformWorkspaceStateJSONRaw)
	ws.GET("/state-json/:state_id/resources", s.GetInstallTerraformWorkspaceStateJSONResources)

	// app-scoped build logs (a build serves many installs, so its logs are
	// not install-scoped).
	logs := api.Group("/v1/apps/:app_id/components/:component_id/builds/:build_id")
	logs.Use(s.requireBuildLogStream)

	logs.GET("/logs", s.ReadComponentBuildLogs)
	logs.GET("/logs/tail", s.TailComponentBuildLogs)
	logs.GET("/spans", s.ReadComponentBuildSpans)

	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	// runners
	runners := api.Group("/v1/runners")
	{
		runners.GET("", s.AdminGetAllRunners)
		runners.GET("/details", s.AdminListRunnersDetails)
		runners.POST("/restart", s.AdminRestartRunners)
		runners.POST("/shutdown-processes", s.AdminShutdownAllRunnerProcesses)
		runners.POST("/update-health-check-cron", s.AdminUpdateHealthCheckCron)
		runners.POST("/migrate-cron-emitters", s.AdminMigrateCronEmitters)
		runners.PATCH("/bulk-update", s.AdminBulkUpdateRunners)

		// sandbox management
		runners.GET("/sandbox", s.AdminListSandboxRunners)
		runners.GET("/sandbox/templates", s.AdminGetSandboxTemplates)

		// runner-specific operations
		runner := runners.Group("/:runner_id")
		{
			runner.GET("", s.AdminGetRunner)

			// runner settings
			runner.GET("/settings", s.AdminGetRunnerSettings)
			runner.PATCH("/settings", s.AdminUpdateRunnerSettings)

			// runner lifecycle
			s.POST(runner, "/reprovision", s.AdminReprovisionRunner, apiPkg.APIContextTypeInternal, true)
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

			// runner processes
			runner.GET("/processes", s.AdminListRunnerProcesses)
			runner.GET("/processes/:process_id", s.AdminGetRunnerProcess)
			runner.POST("/processes/:process_id/shutdown", s.AdminShutdownRunnerProcess)

			// trigger specific jobs
			runner.POST("/graceful-shutdown", s.AdminGracefulShutDown)
			runner.POST("/force-shutdown", s.AdminForceShutDown)
			runner.POST("/mng/shutdown-vm", s.AdminMngVMShutDown)
			runner.POST("/noop-job", s.AdminCreateNoopJob)
			runner.POST("/health-check-job", s.AdminCreateHealthCheck)

			// sandbox config management
			runner.GET("/sandbox-configs", s.AdminGetSandboxConfigs)
			runner.PUT("/sandbox-configs", s.AdminUpsertSandboxConfig)
			runner.DELETE("/sandbox-configs/:config_id", s.AdminDeleteSandboxConfig)
			runner.POST("/sandbox-configs/reset", s.AdminResetSandboxConfigs)
			runner.GET("/sandbox-jobs", s.AdminListSandboxJobs)
		}
	}

	// sandbox mode management
	sandboxMode := api.Group("/v1/sandbox-mode")
	{
		signals := sandboxMode.Group("/signals")
		signals.GET("", s.AdminListSandboxSignalConfigs)
		signals.GET("/types", s.AdminListSignalTypes)
		signals.PUT("/:signal_type", s.AdminUpsertSandboxSignalConfig)
		signals.DELETE("/:signal_type", s.AdminDeleteSandboxSignalConfig)
		signals.POST("/reset", s.AdminResetSandboxSignalConfigs)
		signals.POST("/disable-all", s.AdminDisableAllSandboxSignalConfigs)

		runnerJobs := sandboxMode.Group("/runner-jobs")
		runnerJobs.GET("", s.AdminListAllSandboxConfigs)
		runnerJobs.POST("/disable-all", s.AdminDisableAllSandboxConfigs)
	}

	// org-wide runner settings
	orgs := api.Group("/v1/orgs/:org_id")
	{
		orgs.PATCH("/runner-settings", s.AdminUpdateOrgRunnerSettings)
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
	runners.POST("/component-health", s.CreateComponentHealth)
	runners.GET("/install-components", s.GetRunnerInstallComponents)
	runners.PUT("/component-health-context", s.PutComponentHealthContext)
	runners.GET("/component-health-context", s.GetComponentHealthContext)
	runners.GET("", s.GetRunner)
	runners.GET("/jobs", s.GetRunnerJobs)
	runners.GET("/jobs/tail", s.TailRunnerJobs)
	runners.GET("/settings", s.GetRunnerSettings)
	runners.GET("/public-settings", s.GetRunnerPublicSettings)
	runners.POST("/traces", s.OtelWriteTraces)
	runners.POST("/metrics", s.OtelWriteMetrics)
	runners.GET("/jobs/:job_id/plan", s.GetRunnerJobPlanV2)
	runners.GET("/jobs/:job_id", s.GetRunnerJobV2)
	runners.PATCH("/jobs/:job_id", s.UpdateRunnerJobV2)

	// sandbox configs
	runners.GET("/sandbox-configs", s.GetRunnerSandboxConfigs)
	runners.GET("/sandbox-config", s.GetRunnerSandboxConfig)

	// runner process lifecycle
	runners.POST("/processes", s.CreateRunnerProcess)
	runners.GET("/processes/:process_id", s.GetRunnerProcess)
	runners.PATCH("/processes/:process_id", s.UpdateRunnerProcess)
	runners.GET("/processes/:process_id/shutdowns", s.GetRunnerProcessShutdowns)
	runners.POST("/processes/:process_id/shutdowns/:shutdown_id/complete", s.CompleteRunnerProcessShutdown)
	runners.POST("/processes/:process_id/terminating", s.ReportRunnerProcessTerminating)

	runnerJobs := api.Group("/v1/runner-jobs/:runner_job_id")
	s.GET(runnerJobs, "", s.GetRunnerJob, apiPkg.APIContextTypeRunner, true)
	s.PATCH(runnerJobs, "", s.UpdateRunnerJob, apiPkg.APIContextTypeRunner, true)
	s.GET(runnerJobs, "/plan", s.GetRunnerJobPlan, apiPkg.APIContextTypeRunner, true)
	runnerJobs.GET("/composite-plan", s.GetRunnerJobCompositePlan)

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

	// pulumi state
	pulumiState := api.Group("/v1/runners/pulumi-state")
	pulumiState.GET("/:workspace_id", s.GetPulumiState)
	pulumiState.POST("/:workspace_id", s.UpdatePulumiState)

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

func (s *service) RegisterAuthRoutes(api *gin.Engine) error {
	return nil
}

func (s *service) RegisterAdminDashboardRoutes(api *gin.Engine) error {
	return nil
}

func New(params Params) *service {
	return &service{
		RouteRegister: apiPkg.RouteRegister{
			EndpointAudit: params.EndpointAudit,
		},
		cfg:                  params.Cfg,
		l:                    params.L,
		v:                    params.V,
		db:                   params.DB,
		chDB:                 params.CHDB,
		mw:                   params.MW,
		acctClient:           params.AccountClient,
		helpers:              params.Helpers,
		installsHelpers:      params.InstallsHelpers,
		runnerHeartbeatCache: params.RunnerHeartbeatCache,
		heartbeater:          params.Heartbeater,
		kafka:                params.Kafka,
		featuresClient:       params.FeaturesClient,
		temporalClient:       params.TemporalClient,
		runnerJobWake:        params.RunnerJobWake,
		blobSvc:              params.BlobSvc,
		emitterClient:        params.EmitterClient,
		queueClient:          params.QueueClient,
		logStreamCache:       expirable.NewLRU[string, *app.LogStream](logStreamCacheSize, nil, logStreamCacheTTL),
	}
}

func (s *service) RegisterSlackRoutes(api *gin.Engine) error {
	return nil
}

// requireRunnerJobInInstall gates the install-nested runner-job subtree: the
// job named by :runner_job_id must resolve to :install_id through its
// polymorphic owner (see installshelpers.InstallOwnedScope). It also injects the job's
// log_stream_id as a route param so the shared log handlers — which key off
// :log_stream_id — can be reused unchanged under the runner-job path.
func (s *service) requireRunnerJobInInstall(ctx *gin.Context) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get org from context"))
		ctx.Abort()
		return
	}

	var job app.RunnerJob
	res := s.db.WithContext(ctx).
		Select("id", "log_stream_id").
		Where("id = ? AND org_id = ?", ctx.Param("runner_job_id"), orgID).
		Scopes(s.installsHelpers.InstallOwnedScope(ctx.Param("install_id"))).
		First(&job)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "runner job not found"})
		return
	}
	if res.Error != nil {
		ctx.Error(errors.Wrap(res.Error, "unable to resolve runner job"))
		ctx.Abort()
		return
	}

	if job.LogStreamID != nil {
		ctx.Params = append(ctx.Params, gin.Param{Key: "log_stream_id", Value: *job.LogStreamID})
	}

	ctx.Next()
}

// requireTerraformWorkspaceInInstall gates the install-nested terraform
// workspace subtree: the workspace named by :workspace_id must resolve to
// :install_id through its polymorphic owner (install_components /
// install_sandboxes), via the shared installshelpers.InstallOwnedScope predicate.
func (s *service) requireTerraformWorkspaceInInstall(ctx *gin.Context) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get org from context"))
		ctx.Abort()
		return
	}

	var count int64
	res := s.db.WithContext(ctx).Model(&app.TerraformWorkspace{}).
		Where("id = ? AND org_id = ?", ctx.Param("workspace_id"), orgID).
		Scopes(s.installsHelpers.InstallOwnedScope(ctx.Param("install_id"))).
		Count(&count)
	if res.Error != nil {
		ctx.Error(errors.Wrap(res.Error, "unable to resolve terraform workspace"))
		ctx.Abort()
		return
	}
	if count == 0 {
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "terraform workspace not found"})
		return
	}

	ctx.Next()
}

// requireBuildLogStream gates the app-scoped build-log routes: the build named
// by :build_id must belong to :component_id, which must belong to :app_id (all
// in the caller's org). Build logs are app-scoped rather than install-scoped
// because a build serves many installs. On success it injects the build's
// log_stream_id so the shared log handlers can be reused.
func (s *service) requireBuildLogStream(ctx *gin.Context) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get org from context"))
		ctx.Abort()
		return
	}

	buildID := ctx.Param("build_id")
	componentNameOrID := ctx.Param("component_id")

	var build app.ComponentBuild
	res := s.db.WithContext(ctx).
		Model(&app.ComponentBuild{}).
		Select("component_builds.id").
		Joins("JOIN component_config_connections ON component_config_connections.id = component_builds.component_config_connection_id").
		Joins("JOIN components ON components.id = component_config_connections.component_id").
		Where("component_builds.id = ? AND component_builds.org_id = ?", buildID, orgID).
		Where("components.app_id = ?", ctx.Param("app_id")).
		Where(s.db.Where("components.id = ?", componentNameOrID).Or("components.name = ?", componentNameOrID)).
		First(&build)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "component build not found"})
		return
	}
	if res.Error != nil {
		ctx.Error(errors.Wrap(res.Error, "unable to resolve component build"))
		ctx.Abort()
		return
	}

	var logStream app.LogStream
	res = s.db.WithContext(ctx).
		Select("id").
		Where("owner_type = ? AND owner_id = ? AND org_id = ?", "component_builds", buildID, orgID).
		First(&logStream)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "no logs for build"})
		return
	}
	if res.Error != nil {
		ctx.Error(errors.Wrap(res.Error, "unable to resolve build log stream"))
		ctx.Abort()
		return
	}

	ctx.Params = append(ctx.Params, gin.Param{Key: "log_stream_id", Value: logStream.ID})
	ctx.Next()
}
