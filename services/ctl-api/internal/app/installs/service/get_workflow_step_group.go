package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						GetWorkflowStepGroup
// @Summary					get a workflow step group
// @Description.markdown	get_workflow_step_group.md
// @Param					workflow_id		path	string	true	"workflow ID"
// @Param					step_group_id	path	string	true	"step group ID"
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
// @Success					200	{object}	app.WorkflowStepGroup
// @Router					/v1/workflows/{workflow_id}/step-groups/{step_group_id} [GET]
func (s *service) GetWorkflowStepGroup(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get org from context"))
		return
	}

	workflowID := ctx.Param("workflow_id")
	stepGroupID := ctx.Param("step_group_id")

	group, err := s.getWorkflowStepGroup(ctx, org.ID, workflowID, stepGroupID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get workflow step group"))
		return
	}

	ctx.JSON(http.StatusOK, group)
}

func (s *service) getWorkflowStepGroup(ctx *gin.Context, orgID, workflowID, stepGroupID string) (*app.WorkflowStepGroup, error) {
	var group app.WorkflowStepGroup
	res := s.db.WithContext(ctx).
		Where("id = ? AND workflow_id = ? AND org_id = ?", stepGroupID, workflowID, orgID).
		Preload("Steps", func(db *gorm.DB) *gorm.DB {
			return db.Order("group_retry_idx, idx, created_at asc")
		}).
		Preload("Steps.Approval").
		Preload("Steps.Approval.Response").
		First(&group)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get workflow step group")
	}

	return &group, nil
}
