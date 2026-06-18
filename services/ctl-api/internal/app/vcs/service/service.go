package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type Params struct {
	fx.In

	L           *zap.Logger
	DB          *gorm.DB `name:"psql"`
	V           *validator.Validate
	Helpers     *helpers.Helpers
	GhClient    helpers.GithubClient `optional:"true"`
	BlobSvc     blobstore.Service
	QueueClient *queueclient.Client
}

type service struct {
	l           *zap.Logger
	db          *gorm.DB
	v           *validator.Validate
	helpers     *helpers.Helpers
	ghClient    helpers.GithubClient
	blobSvc     blobstore.Service
	queueClient *queueclient.Client
}

var _ api.Service = (*service)(nil)

func (s *service) RegisterPublicRoutes(api *gin.Engine) error {
	// Webhook event receiver (public, no auth required).
	// Registered outside /v1/vcs to avoid route conflict with the legacy /:vcs_connection_id wildcard.
	api.POST("/v1/vcs/webhooks/:subscription_id/events", s.WriteWebhookEvent)

	// vcs connections
	vcs := api.Group("/v1/vcs")
	{
		vcs.POST("/connection-callback", s.CreateConnectionCallback)

		// Legacy webhook event receiver (per-connection)
		vcs.POST("/:vcs_connection_id/events", s.WriteEvent)

		connections := vcs.Group("/connections")
		{
			connections.POST("", s.CreateConnection)
			connections.GET("", s.GetConnections)
			connections.GET("/:connection_id", s.GetConnection)
			connections.GET("/:connection_id/branches", s.GetVCSConnectionRepoBranches)
			connections.DELETE("/:connection_id", s.DeleteConnection)
			connections.GET("/:connection_id/check-status", s.CheckConnectionStatus)
			connections.GET("/:connection_id/repos", s.ListConnectionRepos)
			connections.GET("/:connection_id/webhook-subscription", s.GetWebhookSubscription)
			connections.POST("/:connection_id/webhook-subscription", s.CreateWebhookSubscription)
		}
	}
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
	return nil
}

func New(params Params) *service {
	var ghClient helpers.GithubClient
	if params.GhClient != nil {
		ghClient = params.GhClient
	} else {
		ghClient = params.Helpers
	}

	return &service{
		v:           params.V,
		l:           params.L,
		db:          params.DB,
		helpers:     params.Helpers,
		ghClient:    ghClient,
		blobSvc:     params.BlobSvc,
		queueClient: params.QueueClient,
	}
}

func (s *service) RegisterSlackRoutes(api *gin.Engine) error {
	return nil
}
