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
	api.GET("/v1/general/current-user", s.GetCurrentUser)
	api.GET("/v1/general/cli-config", s.GetCLIConfig)
	api.GET("/v1/general/cloud-platform/:cloud_platform/regions", s.GetCloudPlatformRegions)
	api.GET("/v1/general/config-schema", s.GetConfigSchema)
	api.POST("/v1/general/waitlist", s.CreateWaitlist)

	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	// manage canaries
	api.POST("/v1/general/provision-canary", s.ProvisionCanary)
	api.POST("/v1/general/deprovision-canary", s.DeprovisionCanary)
	api.POST("/v1/general/start-canary-cron", s.StartCanaryCron)
	api.POST("/v1/general/stop-canary-cron", s.StopCanaryCron)
	api.POST("/v1/general/canary-user", s.CreateCanaryUser)

	// manage infra tests
	api.POST("/v1/general/infra-tests", s.InfraTests)
	api.POST("/v1/general/infra-tests/deprovision", s.InfraTestsDeprovision)

	// create users for testing/seeding
	api.POST("/v1/general/integration-user", s.CreateIntegrationUser)
	api.POST("/v1/general/seed-user", s.CreateSeedUser)

	// migrations
	api.GET("/v1/general/migrations", s.GetMigrations)

	// create a customer token
	api.POST("/v1/general/admin-static-token", s.AdminCreateStaticToken)
	api.POST("/v1/general/admin-delete-account", s.AdminDeleteAccount)

	// (re)start EventLoopReconcile
	api.POST("/v1/general/restart-event-loop", s.RestartGeneralEventLoop)
	api.POST("/v1/general/promotion", s.AdminPromotion)
	api.POST("/v1/general/terminate-event-loops", s.AdminTerminateEventLoops)

	api.GET("/v1/general/waitlist", s.AdminGetWaitlist)

	api.POST("/v1/general/seed", s.Seed)
	api.POST("/v1/general/temporal-codec/decode", s.TemporalCodecDecode)

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
