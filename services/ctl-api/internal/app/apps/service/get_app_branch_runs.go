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

// @ID						GetAppBranchRuns
// @Summary				get app branch workflow runs
// @Description			Returns workflow runs for an app branch ordered by creation time (descending)
// @Tags					apps
// @Param					app_id			path	string	true	"app ID"
// @Param					app_branch_id	path	string	true	"app branch ID"
// @Param					offset			query	int		false	"offset of results to return"	Default(0)
// @Param					limit			query	int		false	"limit of results to return"	Default(10)
// @Param					page			query	int		false	"page number of results to return"	Default(0)
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}		app.Workflow
// @Router					/v1/apps/{app_id}/branches/{app_branch_id}/runs [get]
func (s *service) GetAppBranchRuns(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	// Feature flag checks
	if !org.Features[string(app.OrgFeatureAppBranches)] {
		ctx.Error(fmt.Errorf("app branches feature not enabled for this organization"))
		return
	}

	appID := ctx.Param("app_id")
	appBranchID := ctx.Param("app_branch_id")

	// Verify branch exists and belongs to this org/app
	var branch app.AppBranch
	res := s.db.WithContext(ctx).
		Where(app.AppBranch{
			OrgID: org.ID,
			AppID: appID,
		}).
		First(&branch, "id = ?", appBranchID)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to find app branch: %w", res.Error))
		return
	}

	// Get workflows
	workflows, err := s.getAppBranchRuns(ctx, appBranchID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get workflows: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, workflows)
}

func (s *service) getAppBranchRuns(ctx *gin.Context, appBranchID string) ([]app.Workflow, error) {
	var workflows []app.Workflow

	res := s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination).
		Preload("CreatedBy").
		Preload("Steps").
		Preload("Steps.CreatedBy").
		Preload("Steps.Approval").
		Preload("Steps.Approval.Response").
		Preload("AppBranchRuns").
		Preload("AppBranchRuns.AppBranchConfig").
		Where("owner_type = ?", "app_branches").
		Where("owner_id = ?", appBranchID).
		Order("created_at DESC").
		Find(&workflows)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get workflows: %w", res.Error)
	}

	workflows, err := db.HandlePaginatedResponse(ctx, workflows)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}

	return workflows, nil
}
