package service

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/appbundles/transport"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type Params struct {
	fx.In

	DB             *gorm.DB `name:"psql"`
	Store          transport.Store
	Config         *internal.Config
	AppsHelpers    *appshelpers.Helpers
	QueueClient    *queueclient.Client
	FeaturesClient *features.Features
	L              *zap.Logger
}

type service struct {
	db             *gorm.DB
	store          transport.Store
	cfg            *internal.Config
	appsHelpers    *appshelpers.Helpers
	queueClient    *queueclient.Client
	l              *zap.Logger
	featuresClient *features.Features
}

var _ api.Service = (*service)(nil)

func New(params Params) *service {
	return &service{
		db: params.DB, store: params.Store, cfg: params.Config, appsHelpers: params.AppsHelpers, queueClient: params.QueueClient, l: params.L, featuresClient: params.FeaturesClient,
	}
}

// requireBundleExport is the per-org kill switch for the bundle export surface.
func (s *service) requireBundleExport(ctx *gin.Context) bool {
	enabled, err := s.featuresClient.FeatureEnabled(ctx, app.OrgFeatureAppBundleExport)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to check feature: %w", err))
		return false
	}
	if !enabled {
		ctx.Error(features.ErrFeatureNotEnabled(app.OrgFeatureAppBundleExport))
		return false
	}
	return true
}

func (s *service) RegisterPublicRoutes(api *gin.Engine) error {
	group := api.Group("/v1/apps/:app_id/bundles")
	group.POST("", s.CreateBundle)
	group.GET("", s.ListBundles)
	group.GET("/:bundle_id", s.GetBundle)
	group.POST("/:bundle_id/download-grants", s.CreateDownloadGrant)
	return nil
}

func (s *service) RegisterRunnerRoutes(*gin.Engine) error         { return nil }
func (s *service) RegisterInternalRoutes(*gin.Engine) error       { return nil }
func (s *service) RegisterAuthRoutes(*gin.Engine) error           { return nil }
func (s *service) RegisterAdminDashboardRoutes(*gin.Engine) error { return nil }
func (s *service) RegisterSlackRoutes(*gin.Engine) error          { return nil }
