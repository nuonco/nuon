package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetOrgBranches
// @Summary				get all app branches for an org
// @Description			Returns all app branches across every app in the current org.
// @Param					offset						query	int		false	"offset of branches to return"	Default(0)
// @Param					limit						query	int		false	"limit of branches to return"	Default(10)
// @Param					page						query	int		false	"page number of results to return"	Default(0)
// @Tags					apps
// @Accept					json
// @Produce				json
// @Security				APIKey && OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	[]app.AppBranch
// @Router					/v1/branches [get]
func (s *service) GetOrgBranches(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	branches, err := s.getOrgBranches(ctx, org.ID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, branches)
}

func (s *service) getOrgBranches(ctx *gin.Context, orgID string) ([]app.AppBranch, error) {
	branches := make([]app.AppBranch, 0)

	res := s.db.WithContext(ctx).
		Model(&app.AppBranch{}).
		Select(fmt.Sprintf("app_branches.*, "+
			"(SELECT COUNT(*) FROM %s w "+
			"WHERE w.owner_type = 'app_branches' AND w.owner_id = app_branches.id AND w.deleted_at = 0) AS workflow_count",
			(&app.Workflow{}).TableName())).
		Preload("App").
		Scopes(scopes.WithOffsetPagination).
		Where(app.AppBranch{OrgID: orgID}).
		Order("created_at desc").
		Find(&branches)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get org branches: %w", res.Error)
	}

	branches, err := db.HandlePaginatedResponse(ctx, branches)
	if err != nil {
		return nil, fmt.Errorf("unable to get org branches: %w", err)
	}

	if err := s.attachLatestBranchConfigs(ctx, branches); err != nil {
		return nil, fmt.Errorf("unable to get latest branch configs: %w", err)
	}

	if err := s.attachLatestBranchRuns(ctx, branches); err != nil {
		return nil, fmt.Errorf("unable to get latest branch runs: %w", err)
	}

	return branches, nil
}
