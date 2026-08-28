package service

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/customer_managed/transport"
	installshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/releases"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type Params struct {
	fx.In

	DB              *gorm.DB `name:"psql"`
	CHDB            *gorm.DB `name:"ch"`
	Store           transport.Store
	Config          *internal.Config
	AppsHelpers     *appshelpers.Helpers
	InstallsHelpers *installshelpers.Helpers
	QueueClient     *queueclient.Client
	BlobService     blobstore.Service
	Features        *features.Features
	FlowsClient     *flowclient.Client
	AccountClient   *account.Client
	AuthzClient     *authz.Client
	Releases        releases.Core
}

type service struct {
	db              *gorm.DB
	chDB            *gorm.DB
	store           transport.Store
	cfg             *internal.Config
	appsHelpers     *appshelpers.Helpers
	installsHelpers *installshelpers.Helpers
	queueClient     *queueclient.Client
	blobSvc         blobstore.Service
	features        *features.Features
	flowsClient     *flowclient.Client
	accountClient   *account.Client
	authzClient     *authz.Client
	releases        releases.Core
}

var _ api.Service = (*service)(nil)

func New(params Params) *service {
	return &service{
		db: params.DB, chDB: params.CHDB, store: params.Store, cfg: params.Config, appsHelpers: params.AppsHelpers,
		installsHelpers: params.InstallsHelpers, queueClient: params.QueueClient, blobSvc: params.BlobService, features: params.Features,
 		flowsClient:   params.FlowsClient,
		accountClient: params.AccountClient, authzClient: params.AuthzClient, releases: params.Releases,
	}
}

func (s *service) RegisterPublicRoutes(api *gin.Engine) error {
	releases := api.Group("/v1/apps/:app_id/releases")
	releases.POST("", s.CreateRelease)
	releases.POST("/:release_id/packages", s.CreateReleasePackage)
	releases.GET("/:release_id/packages", s.ListReleasePackages)
	packages := api.Group("/v1/release-packages")
	packages.GET("/:package_id", s.GetReleasePackage)
	packages.POST("/:package_id/download-grants", s.CreateReleasePackageDownloadGrant)
	packages.POST("/:package_id/blob-grants", s.CreateReleasePackageBlobGrants)
	api.POST("/v1/customer-managed/installs", s.CreateCustomerManagedInstall)

	group := api.Group("/v1/apps/:app_id/customer-managed-bundles")
	group.POST("", s.CreateBundle)
	group.GET("", s.ListBundles)
	group.GET("/:bundle_id", s.GetBundle)
	group.POST("/:bundle_id/download-grants", s.CreateDownloadGrant)
	group.POST("/:bundle_id/blob-grants", s.CreateBlobGrants)
	api.POST("/v1/install-registrations", s.RegisterInstall)
	api.GET("/v1/installs/:install_id/release-deployments", s.ListReleaseDeployments)
	api.POST("/v1/installs/:install_id/release-updates", s.CreateReleaseUpdate)
	portal := api.Group("/v1/customer-managed/installs/:install_id")
	portal.GET("/releases", require.Route(permissions.KindInstall, permissions.PermissionRead, "install_id"), s.PortalDiscoverReleases)
	portal.GET("/release-updates", require.Route(permissions.KindInstall, permissions.PermissionRead, "install_id"), s.PortalListReleaseUpdates)
	portal.GET("/releases/:release_id", require.Route(permissions.KindInstall, permissions.PermissionRead, "install_id"), s.PortalGetRelease)
	portal.GET("/releases/:release_id/files/content", require.Route(permissions.KindInstall, permissions.PermissionRead, "install_id"), s.PortalGetReleaseFileContent)
	portal.POST("/releases/:release_id/deploy", require.Route(permissions.KindInstall, permissions.PermissionUpdate, "install_id"), s.PortalDeployRelease)
	portal.GET("/release-packages/:package_id", require.Route(permissions.KindInstall, permissions.PermissionRead, "install_id"), s.PortalGetReleasePackage)
	snapshots := api.Group("/v1/installs/:install_id/support-snapshots")
	snapshots.POST("", s.CreateSupportSnapshot)
	snapshots.GET("", s.ListSupportSnapshots)
	snapshots.GET("/:snapshot_id", s.GetSupportSnapshot)
	return nil
}

func (s *service) RegisterRunnerRoutes(*gin.Engine) error   { return nil }
func (s *service) RegisterInternalRoutes(*gin.Engine) error { return nil }
func (s *service) RegisterAuthRoutes(api *gin.Engine) error {
	return nil
}
func (s *service) RegisterAdminDashboardRoutes(*gin.Engine) error { return nil }
func (s *service) RegisterSlackRoutes(*gin.Engine) error          { return nil }
