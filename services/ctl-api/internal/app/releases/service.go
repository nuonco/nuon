package releases

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
)

type Core interface {
	GetReleaseCore(context.Context, string, string, string) (*app.AppRelease, error)
	GetReleaseFileContentCore(context.Context, string, string, string, string) (*ReleaseFileContentResponse, error)
	ActiveBuildForConnection(context.Context, string, string, app.ComponentConfigConnection) (app.ComponentBuild, error)
}

type Params struct {
	fx.In

	DB          *gorm.DB `name:"psql"`
	AppsHelpers *appshelpers.Helpers
	BlobService blobstore.Service
	Features    *features.Features
}

type service struct {
	db          *gorm.DB
	appsHelpers *appshelpers.Helpers
	blobSvc     blobstore.Service
	features    *features.Features
}

var _ api.Service = (*service)(nil)

func New(params Params) *service {
	return &service{
		db:          params.DB,
		appsHelpers: params.AppsHelpers,
		blobSvc:     params.BlobService,
		features:    params.Features,
	}
}

func NewCore(params Params) Core {
	return New(params)
}

func (s *service) RegisterPublicRoutes(api *gin.Engine) error {
	releases := api.Group("/v1/apps/:app_id/releases")
	releases.GET("", s.ListReleases)
	releases.GET("/:release_id", s.GetRelease)
	releases.GET("/:release_id/files/content", s.GetReleaseFileContent)
	return nil
}

func (s *service) RegisterRunnerRoutes(*gin.Engine) error         { return nil }
func (s *service) RegisterInternalRoutes(*gin.Engine) error       { return nil }
func (s *service) RegisterAuthRoutes(*gin.Engine) error           { return nil }
func (s *service) RegisterAdminDashboardRoutes(*gin.Engine) error { return nil }
func (s *service) RegisterSlackRoutes(*gin.Engine) error          { return nil }
