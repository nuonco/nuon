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
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	orgshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
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

	TemporalCodecGzip         converter.PayloadCodec `name:"gzip"`
	TemporalCodecLargePayload converter.PayloadCodec `name:"largepayload"`
	TemporalCodecS3Payload    converter.PayloadCodec `name:"s3payload"`
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
	codecs         []converter.PayloadCodec
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

func (s *service) RegisterAdminDashboardRoutes(api *gin.Engine) error {
	// Serve static assets
	api.Static("/assets", "./internal/app/admin-dashboard/assets")

	// Register routes - templ components will be rendered directly in handlers
	api.GET("/", s.Index)
	api.GET("/orgs", s.Orgs)
	api.GET("/orgs/table", s.OrgsTable)
	api.GET("/orgs/:id", s.OrgDetail)
	api.GET("/orgs/:id/status", s.OrgStatus)
	api.POST("/orgs/:id/tags", s.UpdateOrgTags)
	api.POST("/orgs/:id/tags/remove/:tag", s.RemoveSingleTag)
	api.POST("/orgs/:id/support-users/add", s.AddSupportUsers)
	api.GET("/orgs/:id/installs/table", s.InstallsTable)

	// Accounts routes
	api.GET("/accounts", s.Accounts)
	api.GET("/accounts/table", s.AccountsTable)
	api.GET("/accounts/:id", s.AccountDetail)
	api.GET("/accounts/:id/installs/table", s.AccountInstallsTable)
	api.GET("/accounts/:id/audit-logs/table", s.AccountAuditLogsTable)

	// Global installs routes
	api.GET("/installs", s.Installs)
	api.GET("/installs/table", s.InstallsTableGlobal)

	// Install detail routes
	api.GET("/installs/:id", s.InstallDetail)
	api.GET("/installs/:id/status/runner", s.InstallRunnerStatus)
	api.GET("/installs/:id/status/sandbox", s.InstallSandboxStatus)
	api.GET("/installs/:id/status/component", s.InstallComponentStatus)
	api.GET("/installs/:id/active-deployments/table", s.InstallActiveDeploymentsTable)
	api.GET("/installs/:id/activity/table", s.InstallActivityTable)
	api.GET("/installs/:id/status/drift", s.InstallDriftStatus)
	api.GET("/installs/:id/workflows/table", s.InstallWorkflowsTable)

	// Workflow routes
	api.GET("/workflows", s.Workflows)
	api.GET("/workflows/table", s.WorkflowsTable)
	api.GET("/workflows/:workflow_id", s.WorkflowDetail)

	// Log stream routes
	api.GET("/log-streams", s.LogStreamViewer)
	api.GET("/log-streams/:log_stream_id", s.LogStreamDetail)
	api.GET("/log-streams/:log_stream_id/logs/table", s.LogStreamLogsTable)

	// Queue routes
	api.GET("/queues", s.Queues)
	api.GET("/queues/table", s.QueuesTable)
	api.GET("/queues/:id", s.QueueDetail)
	api.GET("/queues/:id/emitters/table", s.QueueEmittersTable)
	api.GET("/queues/:id/signals/table", s.QueueSignalsTable)
	api.GET("/queues/:id/signals/:signal_id", s.QueueSignalDetail)
	api.GET("/queues/:id/emitters/:emitter_id", s.QueueEmitterDetail)
	api.POST("/queues/:id/restart", s.RestartQueue)

	// Temporal workflow viewer
	api.GET("/temporal-workflows", s.TemporalWorkflowViewer)

	// Queue signals (global view)
	api.GET("/queue-signals", s.QueueSignals)
	api.GET("/queue-signals/table", s.QueueSignalsGlobalTable)
	api.GET("/queue-signals/signal-type-options", s.QueueSignalTypeOptions)

	// Signal catalog
	api.GET("/signal-catalog", s.SignalCatalog)
	api.GET("/signal-catalog/:signal_type", s.SignalCatalogDetail)

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
		codecs: []converter.PayloadCodec{
			params.TemporalCodecGzip,
			params.TemporalCodecLargePayload,
			params.TemporalCodecS3Payload,
		},
	}

	s.l.Info("admin-dashboard service initialized")
	return s, nil
}
