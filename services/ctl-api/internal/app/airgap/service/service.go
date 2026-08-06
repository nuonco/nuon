package service

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/airgap/transport"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type Params struct {
	fx.In

	DB          *gorm.DB `name:"psql"`
	Store       transport.Store
	Config      *internal.Config
	AppsHelpers *appshelpers.Helpers
	QueueClient *queueclient.Client
}

type service struct {
	db          *gorm.DB
	store       transport.Store
	cfg         *internal.Config
	appsHelpers *appshelpers.Helpers
	queueClient *queueclient.Client
}

var _ api.Service = (*service)(nil)

func New(params Params) *service {
	return &service{db: params.DB, store: params.Store, cfg: params.Config, appsHelpers: params.AppsHelpers, queueClient: params.QueueClient}
}

func (s *service) RegisterPublicRoutes(api *gin.Engine) error {
	group := api.Group("/v1/apps/:app_id/airgap-bundles")
	group.POST("", s.CreateBundle)
	group.GET("", s.ListBundles)
	group.GET("/:bundle_id", s.GetBundle)
	group.POST("/:bundle_id/download-grants", s.CreateDownloadGrant)
	group.POST("/:bundle_id/installs", s.CreateAirgapInstall)
	group.GET("/:bundle_id/installs", s.ListAirgapInstalls)
	return nil
}

func (s *service) RegisterRunnerRoutes(*gin.Engine) error         { return nil }
func (s *service) RegisterInternalRoutes(*gin.Engine) error       { return nil }
func (s *service) RegisterAuthRoutes(*gin.Engine) error           { return nil }
func (s *service) RegisterAdminDashboardRoutes(*gin.Engine) error { return nil }
func (s *service) RegisterSlackRoutes(*gin.Engine) error          { return nil }
