package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetWorkflows
// @Summary					get workflows
// @Description.markdown	get_workflows.md
// @Param					install_id					path	string	true	"install ID"
// @Param					offset						query	int		false	"offset of results to return"	Default(0)
// @Param					limit						query	int		false	"limit of results to return"	Default(10)
// @Param					page						query	int		false	"page number of results to return"	Default(0)
// @Param					x-nuon-pagination-enabled	header	bool	false	"Enable pagination"
// @Tags					installs
// @Accept					json
// @Produce					json
// @Security				APIKey
// @Security				OrgID
// @Failure					400	{object}	stderr.ErrResponse
// @Failure					401	{object}	stderr.ErrResponse
// @Failure					403	{object}	stderr.ErrResponse
// @Failure					404	{object}	stderr.ErrResponse
// @Failure					500	{object}	stderr.ErrResponse
// @Success					200	{array}		app.Workflow
// @Router					/v1/installs/{install_id}/workflows [GET]
func (s *service) GetWorkflows(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	workflows, err := s.getWorkflows(ctx, installID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get workflows"))
		return
	}

	ctx.JSON(http.StatusOK, workflows)
}

func (s *service) getWorkflows(ctx *gin.Context, installID string) ([]app.Workflow, error) {
	var workflows []app.Workflow
	res := s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination).
		Preload("CreatedBy").
		Preload("Steps").
		Preload("Steps.CreatedBy").
		Preload("Steps.Approval").
		Preload("Steps.Approval.Response").
		Where("owner_id = ?", installID).
		Order("created_at desc").
		Find(&workflows)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get workflow runs: %w", res.Error)
	}

	workflows, err := db.HandlePaginatedResponse(ctx, workflows)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}

	return workflows, nil
}
