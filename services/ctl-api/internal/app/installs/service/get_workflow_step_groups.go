package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						GetWorkflowStepGroups
// @Summary					get all step groups for a workflow
// @Description.markdown	get_workflow_step_groups.md
// @Param					workflow_id	path	string	true	"workflow ID"
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
// @Success					200	{array}		app.WorkflowStepGroup
// @Router					/v1/workflows/{workflow_id}/step-groups [GET]
func (s *service) GetWorkflowStepGroups(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get org from context"))
		return
	}

	workflowID := ctx.Param("workflow_id")

	groups, err := s.getWorkflowStepGroups(ctx, org.ID, workflowID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get workflow step groups"))
		return
	}

	ctx.JSON(http.StatusOK, groups)
}

func (s *service) getWorkflowStepGroups(ctx *gin.Context, orgID, workflowID string) ([]app.WorkflowStepGroup, error) {
	var groups []app.WorkflowStepGroup

	res := s.db.WithContext(ctx).
		Where(app.WorkflowStepGroup{
			WorkflowID: workflowID,
			OrgID:      orgID,
		}).
		Preload("Steps", func(db *gorm.DB) *gorm.DB {
			return db.Order("group_retry_idx, idx, created_at asc")
		}).
		Preload("Steps.Approval", func(db *gorm.DB) *gorm.DB {
			return db.Omit("contents")
		}).
		Preload("Steps.Approval.Response").
		Order("group_idx asc").
		Find(&groups)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get workflow step groups")
	}

	return groups, nil
}
