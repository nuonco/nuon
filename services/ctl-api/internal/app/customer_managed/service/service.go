package service

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/customer_managed/transport"
	installshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
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
}

var _ api.Service = (*service)(nil)

func New(params Params) *service {
	return &service{
		db: params.DB, chDB: params.CHDB, store: params.Store, cfg: params.Config, appsHelpers: params.AppsHelpers,
		installsHelpers: params.InstallsHelpers, queueClient: params.QueueClient, blobSvc: params.BlobService, features: params.Features,
		flowsClient: params.FlowsClient, accountClient: params.AccountClient, authzClient: params.AuthzClient,
	}
}

func (s *service) RegisterPublicRoutes(api *gin.Engine) error {
	releases := api.Group("/v1/apps/:app_id/releases")
	releases.POST("", s.CreateRelease)
	releases.GET("", s.ListReleases)
	releases.GET("/:release_id", s.GetRelease)
	releases.GET("/:release_id/files/content", s.GetReleaseFileContent)
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
	agent := api.Group("/v1/customer-managed/installs/:install_id")
	agent.GET("/releases", require.Route(permissions.KindInstall, permissions.PermissionRead, "install_id"), s.AgentDiscoverReleases)
	agent.GET("/release-updates", require.Route(permissions.KindInstall, permissions.PermissionRead, "install_id"), s.AgentListReleaseUpdates)
	agent.GET("/releases/:release_id", require.Route(permissions.KindInstall, permissions.PermissionRead, "install_id"), s.AgentGetRelease)
	agent.GET("/releases/:release_id/files/content", require.Route(permissions.KindInstall, permissions.PermissionRead, "install_id"), s.AgentGetReleaseFileContent)
	agent.GET("/release-packages/:package_id", require.Route(permissions.KindInstall, permissions.PermissionRead, "install_id"), s.AgentGetReleasePackage)
	agent.POST("/release-packages/:package_id/download-grants", require.Route(permissions.KindInstall, permissions.PermissionCreate, "install_id"), s.AgentCreatePackageGrant)
	agent.GET("/workflows", require.Route(permissions.KindInstall, permissions.PermissionRead, "install_id"), s.AgentListWorkflows)
	agent.GET("/workflows/:workflow_id", require.Route(permissions.KindInstall, permissions.PermissionRead, "install_id"), s.AgentGetWorkflow)
	agent.GET("/workflows/:workflow_id/steps/:step_id/logs", require.Route(permissions.KindInstall, permissions.PermissionRead, "install_id"), s.AgentGetWorkflowStepLogs)
	agent.POST("/workflows/:workflow_id/steps/:step_id/retry", require.Route(permissions.KindInstall, permissions.PermissionUpdate, "install_id"), s.AgentRetryWorkflowStep)
	agent.GET("/workflows/:workflow_id/steps/:step_id/approvals/:approval_id/contents", require.Route(permissions.KindInstall, permissions.PermissionRead, "install_id"), s.AgentGetApprovalContents)
	agent.POST("/workflows/:workflow_id/steps/:step_id/approvals/:approval_id/response", require.Route(permissions.KindInstall, permissions.PermissionCreate, "install_id"), s.AgentCreateApprovalResponse)
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
