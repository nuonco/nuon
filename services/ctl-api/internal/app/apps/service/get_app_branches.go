package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
)

// @ID						GetAppBranches
// @Summary				get app branches
// @Description.markdown	get_app_branches.md
// @Param					app_id						path	string	true	"app ID"
// @Param					offset						query	int		false	"offset of branches to return"	Default(0)
// @Param					limit						query	int		false	"limit of branches to return"	Default(10)
// @Param					page						query	int		false	"page number of results to return"	Default(0)
// @Tags					apps
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	[]app.AppBranch
// @Router					/v1/apps/{app_id}/branches [get]
func (s *service) GetAppBranches(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	enabled, err := s.featuresClient.FeatureEnabled(ctx, app.OrgFeatureAppBranches)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to check feature: %w", err))
		return
	}
	if !enabled {
		ctx.Error(features.ErrFeatureNotEnabled(app.OrgFeatureAppBranches))
		return
	}

	appID := ctx.Param("app_id")
	cfgs, err := s.getAppBranches(ctx, org.ID, appID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, cfgs)
}

func (s *service) getAppBranches(ctx *gin.Context, orgID, appID string) ([]app.AppBranch, error) {
	branches := make([]app.AppBranch, 0)

	res := s.db.WithContext(ctx).
		Model(&app.AppBranch{}).
		Select(fmt.Sprintf("app_branches.*, "+
			"(SELECT COUNT(*) FROM %s w "+
			"WHERE w.owner_type = 'app_branches' AND w.owner_id = app_branches.id AND w.deleted_at = 0) AS workflow_count",
			(&app.Workflow{}).TableName())).
		Scopes(scopes.WithOffsetPagination).
		Where(app.AppBranch{
			OrgID: orgID,
			AppID: appID,
		}).
		Order("created_at desc").
		Find(&branches)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app branches: %w", res.Error)
	}

	branches, err := db.HandlePaginatedResponse(ctx, branches)
	if err != nil {
		return nil, fmt.Errorf("unable to get app branches: %w", err)
	}

	return branches, nil
}
