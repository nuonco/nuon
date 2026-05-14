package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.temporal.io/sdk/converter"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/metrics"
	temporal "github.com/nuonco/nuon/pkg/temporal/client"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	generalhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/general/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type service struct {
	apiPkg.RouteRegister
	v              *validator.Validate
	l              *zap.Logger
	db             *gorm.DB
	mw             metrics.Writer
	cfg            *internal.Config
	temporalClient temporal.Client
	authzClient    *authz.Client
	acctClient     *account.Client
	evClient       eventloop.Client
	queueClient    *queueclient.Client
	generalHelpers *generalhelpers.Helpers
	codecs         []converter.PayloadCodec
}

var _ apiPkg.Service = (*service)(nil)

func (s *service) RegisterPublicRoutes(api *gin.Engine) error {
	general := api.Group("/v1/general")
	{
		general.GET("/current-user", s.GetCurrentUser)
		general.GET("/cli-config", s.GetCLIConfig)
		general.GET("/cloud-platform/:cloud_platform/regions", s.GetCloudPlatformRegions)
		general.GET("/config-schema", s.GetConfigSchema)
		general.POST("/waitlist", s.CreateWaitlist)
	}

	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	general := api.Group("/v1/general")
	{
		// manage canaries
		s.POST(general, "/provision-canary", s.ProvisionCanary, apiPkg.APIContextTypeInternal, true)
		s.POST(general, "/deprovision-canary", s.DeprovisionCanary, apiPkg.APIContextTypeInternal, true)
		s.POST(general, "/start-canary-cron", s.StartCanaryCron, apiPkg.APIContextTypeInternal, true)
		s.POST(general, "/stop-canary-cron", s.StopCanaryCron, apiPkg.APIContextTypeInternal, true)
		s.POST(general, "/canary-user", s.CreateCanaryUser, apiPkg.APIContextTypeInternal, true)

		// manage infra tests
		infraTests := general.Group("/infra-tests")
		{
			infraTests.POST("", s.InfraTests)
			infraTests.POST("/deprovision", s.InfraTestsDeprovision)
		}

		// create users for testing/seeding
		general.POST("/integration-user", s.CreateIntegrationUser)
		general.POST("/seed-user", s.CreateSeedUser)

		// migrations
		general.GET("/migrations", s.GetMigrations)

		// admin operations
		general.POST("/admin-static-token", s.AdminCreateStaticToken)
		general.POST("/admin-delete-account", s.AdminDeleteAccount)
		general.POST("/promotion", s.AdminPromotion)
		general.POST("/slack-auto-link", s.AdminSlackAutoLink)
		general.POST("/terminate-event-loops", s.AdminTerminateEventLoops)
		general.GET("/waitlist", s.AdminGetWaitlist)

		// event loop management
		general.POST("/restart-event-loop", s.RestartGeneralEventLoop)

		// seed and utilities
		general.POST("/seed", s.Seed)

		// temporal codec
		general.POST("/temporal-codec/decode", s.TemporalCodecDecode)
	}

	return nil
}

func (s *service) RegisterRunnerRoutes(api *gin.Engine) error {
	api.POST("/v1/general/metrics", s.PublishMetrics)

	return nil
}

func (s *service) RegisterAuthRoutes(api *gin.Engine) error {
	return nil
}

func (s *service) RegisterAdminDashboardRoutes(api *gin.Engine) error {
	return nil
}

type Params struct {
	fx.In

	V              *validator.Validate
	Mw             metrics.Writer
	L              *zap.Logger
	TemporalClient temporal.Client
	Cfg            *internal.Config
	AuthzClient    *authz.Client
	AcctClient     *account.Client
	EvClient       eventloop.Client
	QueueClient    *queueclient.Client
	GeneralHelpers *generalhelpers.Helpers
	DB             *gorm.DB `name:"psql"`
	MW             metrics.Writer

	TemporalCodecGzip         converter.PayloadCodec `name:"gzip"`
	TemporalCodecLargePayload converter.PayloadCodec `name:"largepayload"`
	EndpointAudit             *apiPkg.EndpointAudit
}

func New(params Params) *service {
	return &service{
		RouteRegister: apiPkg.RouteRegister{
			EndpointAudit: params.EndpointAudit,
		},
		l:              params.L,
		v:              params.V,
		mw:             params.Mw,
		db:             params.DB,
		temporalClient: params.TemporalClient,
		cfg:            params.Cfg,
		authzClient:    params.AuthzClient,
		acctClient:     params.AcctClient,
		evClient:       params.EvClient,
		queueClient:    params.QueueClient,
		generalHelpers: params.GeneralHelpers,
		codecs: []converter.PayloadCodec{
			params.TemporalCodecGzip,
			params.TemporalCodecLargePayload,
		},
	}
}

func (s *service) RegisterSlackRoutes(api *gin.Engine) error {
	return nil
}
