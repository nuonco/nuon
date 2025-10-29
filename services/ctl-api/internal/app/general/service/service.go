package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.temporal.io/sdk/converter"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/pkg/metrics"
	temporal "github.com/powertoolsdev/mono/pkg/temporal/client"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/account"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/api"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/authz"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
)

type service struct {
	v              *validator.Validate
	l              *zap.Logger
	db             *gorm.DB
	mw             metrics.Writer
	cfg            *internal.Config
	temporalClient temporal.Client
	authzClient    *authz.Client
	acctClient     *account.Client
	evClient       eventloop.Client
	codecs         []converter.PayloadCodec
}

var _ api.Service = (*service)(nil)

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
		general.POST("/provision-canary", s.ProvisionCanary)
		general.POST("/deprovision-canary", s.DeprovisionCanary)
		general.POST("/start-canary-cron", s.StartCanaryCron)
		general.POST("/stop-canary-cron", s.StopCanaryCron)
		general.POST("/canary-user", s.CreateCanaryUser)

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
	DB             *gorm.DB `name:"psql"`
	MW             metrics.Writer

	TemporalCodecGzip         converter.PayloadCodec `name:"gzip"`
	TemporalCodecLargePayload converter.PayloadCodec `name:"largepayload"`
}

func New(params Params) *service {
	return &service{
		l:              params.L,
		v:              params.V,
		mw:             params.Mw,
		db:             params.DB,
		temporalClient: params.TemporalClient,
		cfg:            params.Cfg,
		authzClient:    params.AuthzClient,
		acctClient:     params.AcctClient,
		evClient:       params.EvClient,
		codecs: []converter.PayloadCodec{
			params.TemporalCodecGzip,
			params.TemporalCodecLargePayload,
		},
	}
}
