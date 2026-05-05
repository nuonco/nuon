package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.temporal.io/sdk/converter"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/metrics"
	temporalclient "github.com/nuonco/nuon/pkg/temporal/client"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	orgshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/querycollector"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	emitterclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/client"
)

type Params struct {
	fx.In
	V              *validator.Validate
	Cfg            *internal.Config
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	MW             metrics.Writer
	L              *zap.Logger
	AppsHelpers    *appshelpers.Helpers
	AcctClient     *account.Client
	AuthzClient    *authz.Client
	OrgsHelpers    *orgshelpers.Helpers
	TemporalClient temporalclient.Client
	QueueClient    *queueclient.Client
	EmitterClient  *emitterclient.Client

	TemporalCodecGzip         converter.PayloadCodec `name:"gzip"`
	TemporalCodecLargePayload converter.PayloadCodec `name:"largepayload"`
	TemporalCodecS3Payload    converter.PayloadCodec `name:"s3payload"`

	QueryCollector *querycollector.Collector
}

type Service struct {
	v              *validator.Validate
	l              *zap.Logger
	db             *gorm.DB
	chDB           *gorm.DB
	mw             metrics.Writer
	cfg            *internal.Config
	appsHelpers    *appshelpers.Helpers
	acctClient     *account.Client
	authzClient    *authz.Client
	orgsHelpers    *orgshelpers.Helpers
	temporalClient temporalclient.Client
	queueClient    *queueclient.Client
	emitterClient  *emitterclient.Client
	codecs         []converter.PayloadCodec
	queryCollector *querycollector.Collector
}

type service = Service

var _ api.Service = (*service)(nil)

func (s *service) RegisterPublicRoutes(api *gin.Engine) error {
	return nil
}

func (s *service) RegisterRunnerRoutes(api *gin.Engine) error {
	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	return nil
}

func (s *service) RegisterAuthRoutes(api *gin.Engine) error {
	return nil
}

func (s *service) setAdminAccountContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to resolve account from X-Nuon-Auth cookie
		token, _ := c.Cookie("X-Nuon-Auth")
		if token != "" {
			var userToken app.Token
			if res := s.db.WithContext(c).Where(&app.Token{Token: token}).First(&userToken); res.Error == nil {
				if acct, err := s.acctClient.FetchAccount(c, userToken.AccountID); err == nil {
					cctx.SetAccountGinContext(c, acct)
					c.Next()
					return
				}
			}
		}

		// Fallback: set account ID on the request context so GORM BeforeCreate hooks
		// can read it via createdByIDFromContext(tx.Statement.Context).
		ctx := cctx.SetAccountIDContext(c.Request.Context(), "admin-dashboard")
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func (s *service) RegisterAdminDashboardRoutes(e *gin.Engine) error {
	// Set admin account context from X-Nuon-Auth cookie
	e.Use(s.setAdminAccountContext())

	// Health check
	e.GET("/api/livez", s.Livez)

	// JSON API endpoints
	api := e.Group("/api")
	{
		// Orgs
		api.GET("/orgs", s.Orgs)
		api.GET("/orgs/table", s.OrgsTable)
		api.GET("/orgs/:id", s.OrgDetail)
		api.GET("/orgs/:id/status", s.OrgStatus)
		api.POST("/orgs/:id/labels", s.AddOrgLabels)
		api.POST("/orgs/:id/labels/remove/:key", s.RemoveOrgLabel)
		api.POST("/orgs/:id/support-users/add", s.ProxyAddSupportUsers)
		api.POST("/orgs/:id/migrate-queues", s.ProxyMigrateQueues)
		api.GET("/orgs/:id/installs", s.InstallsTable)

		// Accounts
		api.GET("/accounts", s.Accounts)
		api.GET("/accounts/table", s.AccountsTable)
		api.GET("/accounts/:id", s.AccountDetail)
		api.GET("/accounts/:id/installs", s.AccountInstallsTable)
		api.GET("/accounts/:id/audit-logs", s.AccountAuditLogsTable)

		// Runner uptime
		api.GET("/runner-uptime", s.RunnerUptime)

		// Runners
		api.GET("/runners/all", s.AllRunners)
		api.GET("/runners/:id", s.RunnerDetail)
		api.PUT("/runners/:id/configs", s.RunnerUpsertConfig)
		api.DELETE("/runners/:id/configs/:job_type", s.RunnerDeleteConfig)
		api.POST("/runners/:id/configs/reset", s.RunnerResetConfigs)

		// Labels
		api.GET("/labels", s.LabelsPage)
		api.GET("/labels/table", s.LabelsTable)

		// Global installs
		api.GET("/installs", s.Installs)
		api.GET("/installs/table", s.InstallsTableGlobal)

		// Install detail
		api.GET("/installs/:id", s.InstallDetail)
		api.GET("/installs/:id/status/runner", s.InstallRunnerStatus)
		api.GET("/installs/:id/status/sandbox", s.InstallSandboxStatus)
		api.GET("/installs/:id/status/component", s.InstallComponentStatus)
		api.GET("/installs/:id/active-deployments", s.InstallActiveDeploymentsTable)
		api.GET("/installs/:id/activity", s.InstallActivityTable)
		api.GET("/installs/:id/status/drift", s.InstallDriftStatus)
		api.GET("/installs/:id/workflows", s.InstallWorkflowsTable)
		api.POST("/installs/:id/labels", s.AddInstallLabel)
		api.POST("/installs/:id/labels/remove/:key", s.RemoveInstallLabel)

		// Workflows
		api.GET("/workflows", s.Workflows)
		api.GET("/workflows/table", s.WorkflowsTable)
		api.GET("/workflows/:workflow_id", s.WorkflowDetail)

		// Log streams
		api.GET("/log-streams", s.LogStreamViewer)
		api.GET("/log-streams/:log_stream_id", s.LogStreamDetail)
		api.GET("/log-streams/:log_stream_id/logs", s.LogStreamLogsTable)

		// Queue routes
		api.GET("/queues", s.Queues)
		api.GET("/queues/table", s.QueuesTable)
		api.GET("/queues/:id", s.QueueDetail)
		api.GET("/queues/:id/emitters", s.QueueEmittersTable)
		api.GET("/queues/:id/signals", s.QueueSignalsTable)
		api.GET("/queues/:id/in-flight-signals", s.QueueInFlightSignalsTable)
		api.GET("/queues/:id/signals/:signal_id", s.QueueSignalDetail)
		api.GET("/queues/:id/signals/:signal_id/graph", s.SignalGraph)
		api.GET("/queues/:id/emitters/:emitter_id", s.QueueEmitterDetail)
		api.POST("/queues/:id/restart", s.RestartQueue)
		api.POST("/queues/:id/force-restart", s.ForceRestartQueue)
		api.POST("/queues/:id/clear", s.ClearQueue)
		api.POST("/queues/:id/signals/:signal_id/direct-execute", s.DirectExecuteSignal)

		// Temporal workflow viewer
		api.GET("/temporal-workflows", s.TemporalWorkflowViewer)

		// Temporal workers
		api.GET("/temporal-workers", s.TemporalWorkers)
		api.GET("/temporal-workers/table", s.TemporalWorkersTable)
		api.GET("/temporal-workers/:namespace", s.TemporalWorkerDetail)

		// Queue signals (global view)
		api.GET("/queue-signals", s.QueueSignals)
		api.GET("/queue-signals/table", s.QueueSignalsGlobalTable)
		api.GET("/queue-signals/signal-type-options", s.QueueSignalTypeOptions)

		// In-flight signals
		api.GET("/in-flight-signals", s.InFlightSignals)
		api.GET("/in-flight-signals/table", s.InFlightSignalsTable)

		// Signal catalog
		api.GET("/signal-catalog", s.SignalCatalog)
		api.GET("/signal-catalog/:signal_type", s.SignalCatalogDetail)

		// Sandbox mode
		api.GET("/sandbox-mode", s.SandboxMode)
		api.GET("/sandbox-mode/runner-jobs", s.SandboxModeRunnerJobsTable)
		api.GET("/sandbox-mode/runner-jobs/rows", s.SandboxModeRunnerJobsRows)
		api.GET("/sandbox-mode/builder", s.SandboxModeBuilder)
		api.GET("/sandbox-mode/signals", s.SandboxModeSignalsTable)
		api.GET("/sandbox-mode/signals/rows", s.SandboxModeSignalRows)
		api.GET("/sandbox-mode/stacks", s.SandboxModeStacksTable)
		api.PUT("/sandbox-mode/signals/:signal_type", s.SandboxModeUpsertSignalConfig)
		api.PUT("/sandbox-mode/runner-jobs/:job_type", s.SandboxModeUpsertRunnerJobConfig)
		api.POST("/sandbox-mode/signals/disable-all", s.SandboxModeDisableAllSignals)
		api.POST("/sandbox-mode/runner-jobs/disable-all", s.SandboxModeDisableAllRunnerJobs)
		api.POST("/sandbox-mode/templates/:template_key/apply", s.SandboxModeApplyFlowTemplate)

		// Queries (dev-only)
		api.GET("/queries", s.Queries)
		api.POST("/queries/clear", s.ClearQueries)

		// Query catalog
		api.GET("/query-catalog", s.QueryCatalogList)
		api.POST("/query-catalog/:query_id/run", s.QueryCatalogRun)
		api.POST("/query-collector/toggle", s.QueryCollectorToggle)

		// General actions
		api.POST("/promote", s.Promote)
		api.POST("/seed", s.ProxySeed)
	}

	// SPA serving (must be AFTER api routes)
	s.registerSPARoutes(e)

	s.l.Info("admin-dashboard routes registered")
	return nil
}

func New(params Params) (*service, error) {
	s := &service{
		cfg:            params.Cfg,
		l:              params.L,
		v:              params.V,
		db:             params.DB,
		chDB:           params.CHDB,
		mw:             params.MW,
		appsHelpers:    params.AppsHelpers,
		acctClient:     params.AcctClient,
		authzClient:    params.AuthzClient,
		orgsHelpers:    params.OrgsHelpers,
		temporalClient: params.TemporalClient,
		queueClient:    params.QueueClient,
		emitterClient:  params.EmitterClient,
		queryCollector: params.QueryCollector,
		codecs: []converter.PayloadCodec{
			params.TemporalCodecGzip,
			params.TemporalCodecLargePayload,
			params.TemporalCodecS3Payload,
		},
	}

	s.l.Info("admin-dashboard service initialized")
	return s, nil
}
