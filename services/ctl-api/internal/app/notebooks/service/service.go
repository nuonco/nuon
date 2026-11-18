package service

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	notebookaccess "github.com/nuonco/nuon/services/ctl-api/internal/app/notebooks/access"
	notebookclient "github.com/nuonco/nuon/services/ctl-api/internal/app/notebooks/client"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type Params struct {
	fx.In

	V              *validator.Validate
	DB             *gorm.DB `name:"psql"`
	L              *zap.Logger
	Cfg            *internal.Config
	InstallHelpers *installhelpers.Helpers
	EndpointAudit  *apiPkg.EndpointAudit
	FeaturesClient *features.Features
	NotebookClient *notebookclient.Client
	QueueClient    *queueclient.Client
}

type service struct {
	apiPkg.RouteRegister
	v              *validator.Validate
	db             *gorm.DB
	l              *zap.Logger
	cfg            *internal.Config
	installHelpers *installhelpers.Helpers
	featuresClient *features.Features
	notebookClient *notebookclient.Client
	queueClient    *queueclient.Client
}

var _ apiPkg.Service = (*service)(nil)

func New(params Params) *service {
	return &service{
		RouteRegister: apiPkg.RouteRegister{
			EndpointAudit: params.EndpointAudit,
		},
		v:              params.V,
		db:             params.DB,
		l:              params.L,
		cfg:            params.Cfg,
		installHelpers: params.InstallHelpers,
		featuresClient: params.FeaturesClient,
		notebookClient: params.NotebookClient,
		queueClient:    params.QueueClient,
	}
}

func (s *service) RegisterPublicRoutes(api *gin.Engine) error {
	installs := api.Group("/v1/installs/:install_id")
	{
		notebooks := installs.Group("/notebooks")
		notebooksAccess := require.Routes(notebooks, permissions.APICtl)
		{
			notebooksAccess.POST("", require.Point(permissions.NotebookCreate, s.notebookInstallTarget), s.CreateNotebook)
			notebooksAccess.GET("", s.notebookListRule(notebookaccess.Notebooks), s.GetNotebooks)
			notebooksAccess.GET("/:notebook_id", require.Point(permissions.NotebookInspect, s.notebookTarget), s.GetNotebook)
			notebooksAccess.PATCH("/:notebook_id", require.Point(permissions.NotebookUpdate, s.notebookTarget), s.UpdateNotebook)
			notebooksAccess.DELETE("/:notebook_id", require.Point(permissions.NotebookDelete, s.notebookTarget), s.DeleteNotebook)

			cells := require.Routes(notebooks.Group("/:notebook_id/cells"), permissions.APICtl)
			{
				cells.POST("", require.Point(permissions.NotebookUpdate, s.notebookTarget), s.CreateCell)
				cells.PATCH("/:cell_id", require.Point(permissions.NotebookUpdate, s.notebookCellTarget), s.UpdateCell)
				cells.DELETE("/:cell_id", require.Point(permissions.NotebookDelete, s.notebookCellTarget), s.DeleteCell)
				cells.PUT("/reorder", require.Point(permissions.NotebookUpdate, s.notebookReorderTarget), s.ReorderCells)
				cells.POST("/:cell_id/runs", require.Point(permissions.NotebookExecute, s.notebookExecutionTarget), s.RunCell)
				cells.GET("/:cell_id/runs", s.notebookListRule(notebookaccess.Runs), s.GetCellRuns)
			}
			notebooksAccess.GET("/:notebook_id/runs/:run_id", require.Point(permissions.NotebookInspect, s.notebookRunTarget), s.GetCellRun)
		}
	}

	return nil
}

func (s *service) RegisterAuthRoutes(api *gin.Engine) error           { return nil }
func (s *service) RegisterInternalRoutes(api *gin.Engine) error       { return nil }
func (s *service) RegisterRunnerRoutes(api *gin.Engine) error         { return nil }
func (s *service) RegisterAdminDashboardRoutes(api *gin.Engine) error { return nil }
func (s *service) RegisterSlackRoutes(api *gin.Engine) error          { return nil }

// gate verifies the notebooks feature is enabled and resolves the org +
// install for the request. All notebook handlers begin with it.
func (s *service) gate(ctx *gin.Context) (*app.Org, *app.Install, error) {
	enabled, err := s.featuresClient.FeatureEnabled(ctx, app.OrgFeatureNotebooks)
	if err != nil || !enabled {
		return nil, nil, fmt.Errorf("notebooks feature is not enabled")
	}

	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	install, err := s.installHelpers.GetInstall(ctx, org.ID, ctx.Param("install_id"))
	if err != nil {
		return nil, nil, fmt.Errorf("install not found: %w", err)
	}

	return org, install, nil
}

// getNotebook loads a notebook scoped to the org + install in the request.
func (s *service) getNotebook(ctx *gin.Context, orgID, installID, notebookID string) (*app.Notebook, error) {
	var nb app.Notebook
	res := s.db.WithContext(ctx).
		Where(app.Notebook{OrgID: orgID, InstallID: installID}).
		First(&nb, "id = ?", notebookID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get notebook: %w", res.Error)
	}
	return &nb, nil
}
