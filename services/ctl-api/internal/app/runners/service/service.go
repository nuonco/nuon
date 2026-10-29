package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"go.opentelemetry.io/otel/metric"
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
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/heartbeater"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/kafka"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	emitterclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/telemetrytoken"
)

type Params struct {
	fx.In

	V                    *validator.Validate
	Cfg                  *internal.Config
	DB                   *gorm.DB `name:"psql"`
	CHDB                 *gorm.DB `name:"ch"`
	MW                   metrics.Writer
	MeterProvider        metric.MeterProvider `optional:"true"`
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
	v                      *validator.Validate
	l                      *zap.Logger
	db                     *gorm.DB
	chDB                   *gorm.DB
	mw                     metrics.Writer
	tailMetrics            *runnerJobTailMetrics
	executionResults       metric.Int64Counter
	meterProvider          metric.MeterProvider
	cfg                    *internal.Config
	acctClient             *account.Client
	helpers                *helpers.Helpers
	installsHelpers        *installshelpers.Helpers
	runnerHeartbeatCache   *RunnerHeartbeatCache
	heartbeater            *heartbeater.Heartbeater
	kafka                  *kafka.Producer
	featuresClient         *features.Features
	temporalClient         temporalclient.Client
	runnerJobWake          *RunnerJobWakeRegistry
	blobSvc                blobstore.Service
	emitterClient          *emitterclient.Client
	queueClient            *queueclient.Client
	telemetryTokenIssuer   *telemetrytoken.Issuer
	telemetryRelayEndpoint string
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
	api.GET("/v1/runner-jobs", s.ListRunnerJobsCtlAPI)
	api.GET("/v1/runner-jobs/:runner_job_id/plan", s.GetRunnerJobPlanPublic)
	api.GET("/v1/runner-jobs/:runner_job_id/composite-plan", s.GetRunnerJobCompositePlan)
	api.POST("/v1/runner-jobs/:runner_job_id/cancel", s.CancelRunnerJob)
	api.GET("/v1/runner-jobs/:runner_job_id", s.GetRunnerJobPublic)

	runners := api.Group("/v1/runners/:runner_id", require.RunnerInOrg(s.db))
	{
		runners.GET("", s.GetRunnerCtlAPI)
		runners.GET("/connected", s.GetRunnerConnectStatus)
		runners.GET("/jobs", s.GetRunnerJobsCtlAPI)
		runners.GET("/recent-health-checks", s.GetRunnerRecentHealthChecks)
		runners.GET("/latest-heart-beat", s.GetRunnerLatestHeartBeat)
		runners.GET("/heart-beats/latest", s.GetLatestRunnerHeartBeatFromView)
		runners.GET("/card-details", s.GetRunnerCardDetails)

		runners.GET("/processes", s.ListRunnerProcesses)
		runners.GET("/processes/current", s.GetCurrentRunnerProcesses)
		runners.GET("/processes/:process_id", s.GetRunnerProcessPublic)
		runners.GET("/processes/:process_id/heart-beats/latest", s.GetProcessLatestHeartBeat)
		runners.POST("/processes/:process_id/shutdown", s.ShutdownRunnerProcess)

		runners.POST("/graceful-shutdown", s.GracefulShutDown)
		runners.POST("/force-shutdown", s.ForceShutDown)
		runners.POST("/mng/shutdown-vm", s.MngVMShutDown)
		runners.POST("/mng/shutdown", s.MngShutDown)
		runners.POST("/mng/update", s.MngUpdate)
		runners.POST("/mng/restart", s.MngRestart)
		runners.POST("/prune-tokens", s.PruneTokens)

		runners.GET("/settings", s.GetRunnerSettingsPublic)
		runners.PATCH("/settings", s.UpdateRunnerSettings)
	}

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
		runners.POST("/toggle-cron-emitters", s.AdminToggleCronEmitters)

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
		runnerJobs.PUT("/:job_type", s.AdminUpsertSandboxConfig)
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
	s.tailMetrics = newRunnerJobTailMetrics(s.meterProvider)
	s.executionResults = newRunnerJobExecutionResults(s.meterProvider)
	// In-job APIs: root runner has org all; job SA has install-scoped grants only.
	org := require.OrgOrInstallRoute(api)
	org.POST("/v1/telemetry/access-token", s.CreateTelemetryAccessToken)

	runners := api.Group("/v1/runners/:runner_id")
	runnerRoutes := require.OrgRoute(runners)
	runnerRoutes.POST("/health-checks", s.CreateRunnerHealthCheck)
	runnerRoutes.POST("/heart-beats", s.CreateRunnerHeartBeat)
	runnerRoutes.POST("/component-health", s.CreateComponentHealth)
	runnerRoutes.GET("/install-components", s.GetRunnerInstallComponents)
	runnerRoutes.PUT("/component-health-context", s.PutComponentHealthContext)
	runnerRoutes.GET("/component-health-context", s.GetComponentHealthContext)
	runnerRoutes.GET("", s.GetRunner)
	runnerRoutes.GET("/jobs", s.GetRunnerJobs)
	runnerRoutes.GET("/jobs/tail", s.TailRunnerJobs)
	runnerRoutes.GET("/settings", s.GetRunnerSettings)
	runners.GET("/public-settings", s.GetRunnerPublicSettings)
	runnerRoutes.POST("/traces", s.OtelWriteTraces)
	runnerRoutes.POST("/metrics", s.OtelWriteMetrics)
	runnerRoutes.POST("/jobs/:job_id/token", s.ClaimRunnerJobToken)
	// Job SA has object-scoped runner-job grants, not org-level read — these must use ResourceRoute.
	jobByRunner := require.ResourceRoute(runners, permissions.KindRunnerJob, "job_id")
	jobByRunner.GET("/jobs/:job_id", s.GetRunnerJobV2)
	jobByRunner.GET("/jobs/:job_id/plan", s.GetRunnerJobPlanV2)
	jobByRunner.PATCH("/jobs/:job_id", s.UpdateRunnerJobV2)

	// sandbox configs — job token may load these during exec
	sandboxConfigs := require.OrgOrInstallRoute(runners)
	sandboxConfigs.GET("/sandbox-configs", s.GetRunnerSandboxConfigs)
	sandboxConfigs.GET("/sandbox-config", s.GetRunnerSandboxConfig)

	// runner process lifecycle
	runnerRoutes.POST("/processes", s.CreateRunnerProcess)
	runnerRoutes.GET("/processes/:process_id", s.GetRunnerProcess)
	runnerRoutes.PATCH("/processes/:process_id", s.UpdateRunnerProcess)
	// When public, register outside OrgRoute so unauthenticated upgrade-path
	// runners can poll; public middleware allowlists the path when the flag is on.
	// cfg may be nil in route-registration tests that construct a bare service.
	if s.cfg != nil && s.cfg.PublicRunnerProcessShutdowns {
		runners.GET("/processes/:process_id/shutdowns", s.GetRunnerProcessShutdowns)
	} else {
		runnerRoutes.GET("/processes/:process_id/shutdowns", s.GetRunnerProcessShutdowns)
	}
	runnerRoutes.POST("/processes/:process_id/shutdowns/:shutdown_id/complete", s.CompleteRunnerProcessShutdown)
	runnerRoutes.POST("/processes/:process_id/terminating", s.ReportRunnerProcessTerminating)

	// Reads and writes share ResourceRoute so the job SA's org:runner-job/<id> grant applies.
	jobRoutes := require.ResourceRoute(api, permissions.KindRunnerJob, "runner_job_id")
	jobRoutes.GET("/v1/runner-jobs/:runner_job_id", s.GetRunnerJob)
	jobRoutes.GET("/v1/runner-jobs/:runner_job_id/plan", s.GetRunnerJobPlan)
	jobRoutes.GET("/v1/runner-jobs/:runner_job_id/composite-plan", s.GetRunnerJobCompositePlan)
	jobRoutes.GET("/v1/runner-jobs/:runner_job_id/executions", s.GetRunnerJobExecutions)
	jobRoutes.GET("/v1/runner-jobs/:runner_job_id/executions/:runner_job_execution_id", s.GetRunnerJobExecution)
	jobRoutes.PATCH("/v1/runner-jobs/:runner_job_id", s.UpdateRunnerJob)
	jobRoutes.POST("/v1/runner-jobs/:runner_job_id/executions", s.CreateRunnerJobExecution)
	jobRoutes.PATCH("/v1/runner-jobs/:runner_job_id/executions/:runner_job_execution_id", s.UpdateRunnerJobExecution)
	jobRoutes.POST("/v1/runner-jobs/:runner_job_id/executions/:runner_job_execution_id/result", s.CreateRunnerJobExecutionResult)
	jobRoutes.POST("/v1/runner-jobs/:runner_job_id/executions/:runner_job_execution_id/outputs", s.CreateRunnerJobExecutionOutputs)
	s.EndpointAudit.Add("GET", string(apiPkg.APIContextTypeRunner), "/v1/runner-jobs/:runner_job_id")
	s.EndpointAudit.Add("PATCH", string(apiPkg.APIContextTypeRunner), "/v1/runner-jobs/:runner_job_id")
	s.EndpointAudit.Add("GET", string(apiPkg.APIContextTypeRunner), "/v1/runner-jobs/:runner_job_id/plan")

	// Terraform backend
	tfBackend := org.Group("/v1/terraform-backend")
	tfBackend.GET("", s.GetTerraformCurrentStateData)
	tfBackend.POST("", s.UpdateTerraformState)
	tfBackend.DELETE("", s.DeleteTerraformState)

	// pulumi state
	pulumiState := org.Group("/v1/runners/pulumi-state")
	pulumiState.GET("/:workspace_id", s.GetPulumiState)
	pulumiState.POST("/:workspace_id", s.UpdatePulumiState)

	// terraform workspaces
	tfWorkspaces := org.Group("/v1/terraform-workspaces")
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
	org.GET(helmReleasePath+":namespace", s.GetHelmReleases)
	org.GET(helmReleasePath+":namespace/:key", s.GetHelmRelease)
	org.GET(helmReleasePath+":namespace/query", s.QueryHelmRelease)
	org.POST(helmReleasePath+":namespace/:key", s.CreateHelmRelease)
	org.PUT(helmReleasePath+":namespace/:key", s.UpdateHelmRelease)
	org.DELETE(helmReleasePath+":namespace/:key", s.DeleteHelmRelease)

	// TODO(jm): these will be moved to the otel namespace
	org.POST("/v1/log-streams/:log_stream_id/logs", s.LogStreamWriteLogs)

	// installs
	installs := org.Group("/v1/installs")
	installs.GET("/:install_id/:component_id/last-active-plan", s.GetInstallComponenetLastActivePlan)

	return nil
}

func (s *service) RegisterAuthRoutes(api *gin.Engine) error {
	return nil
}

func (s *service) RegisterAdminDashboardRoutes(api *gin.Engine) error {
	return nil
}

func New(params Params) (*service, error) {
	telemetryTokenIssuer, err := telemetrytoken.New(params.Cfg)
	if err != nil {
		return nil, fmt.Errorf("initialize telemetry token issuer: %w", err)
	}
	telemetryRelayEndpoint, err := newTelemetryRelayEndpoint(params.Cfg, telemetryTokenIssuer)
	if err != nil {
		return nil, fmt.Errorf("invalid telemetry relay configuration: %w", err)
	}

	return &service{
		RouteRegister: apiPkg.RouteRegister{
			EndpointAudit: params.EndpointAudit,
		},
		cfg:                    params.Cfg,
		l:                      params.L,
		v:                      params.V,
		db:                     params.DB,
		chDB:                   params.CHDB,
		mw:                     params.MW,
		meterProvider:          params.MeterProvider,
		acctClient:             params.AccountClient,
		helpers:                params.Helpers,
		installsHelpers:        params.InstallsHelpers,
		runnerHeartbeatCache:   params.RunnerHeartbeatCache,
		heartbeater:            params.Heartbeater,
		kafka:                  params.Kafka,
		featuresClient:         params.FeaturesClient,
		temporalClient:         params.TemporalClient,
		runnerJobWake:          params.RunnerJobWake,
		blobSvc:                params.BlobSvc,
		emitterClient:          params.EmitterClient,
		queueClient:            params.QueueClient,
		telemetryTokenIssuer:   telemetryTokenIssuer,
		telemetryRelayEndpoint: telemetryRelayEndpoint,
		logStreamCache:         expirable.NewLRU[string, *app.LogStream](logStreamCacheSize, nil, logStreamCacheTTL),
	}, nil
}

func (s *service) RegisterSlackRoutes(api *gin.Engine) error {
	return nil
}
