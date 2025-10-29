package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

// @ID									GetWorkflowStepApproval
// @Summary								get an workflow step approval
// @Description.markdown				get_workflow_step_approval.md
// @Param	workflow_id			path	string	true	"workflow id"
// @Param	step_id	path	string	true	"step id"
// @Param	approval_id					path	string	true	"approval id"
// @Tags								installs
// @Accept								json
// @Produce								json
// @Security							APIKey
// @Security							OrgID
// @Failure								400	{object}	stderr.ErrResponse
// @Failure								401	{object}	stderr.ErrResponse
// @Failure								403	{object}	stderr.ErrResponse
// @Failure								404	{object}	stderr.ErrResponse
// @Failure								500	{object}	stderr.ErrResponse
// @Success								200	{object}		app.WorkflowStepApproval
// @Router 								/v1/workflows/{workflow_id}/steps/{step_id}/approvals/{approval_id} [GET]
func (s *service) GetWorkflowStepApproval(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get org from context"))
		return
	}

	workflowID := ctx.Param("workflow_id")
	stepID := ctx.Param("step_id")
	approvalID := ctx.Param("approval_id")

	_, err = s.getWorkflowStep(ctx, workflowID, stepID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get workflow step"))
		return
	}

	approval, err := s.getWorkflowStepApproval(ctx, org.ID, approvalID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get workflow step approval"))
		return
	}

	ctx.JSON(http.StatusOK, approval)
}

// TODO: Remove. Deprecated.
// @ID									GetInstallWorkflowStepApproval
// @Summary								get an install workflow step approval
// @Description.markdown				get_workflow_step_approval.md
// @Param	install_workflow_id			path	string	true	"workflow id"
// @Param	install_workflow_step_id	path	string	true	"step id"
// @Param	approval_id					path	string	true	"approval id"
// @Tags								installs
// @Accept								json
// @Produce								json
// @Security							APIKey
// @Security							OrgID
// @Failure								400	{object}	stderr.ErrResponse
// @Failure								401	{object}	stderr.ErrResponse
// @Failure								403	{object}	stderr.ErrResponse
// @Failure								404	{object}	stderr.ErrResponse
// @Failure								500	{object}	stderr.ErrResponse
// @Success								200	{object}		app.WorkflowStepApproval
// @Router 								/v1/install-workflows/{install_workflow_id}/steps/{install_workflow_step_id}/approvals/{approval_id} [GET]
// @Deprecated
func (s *service) GetInstallWorkflowStepApproval(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get org from context"))
		return
	}

	workflowID := ctx.Param("install_workflow_id")
	stepID := ctx.Param("install_workflow_step_id")
	approvalID := ctx.Param("approval_id")

	_, err = s.getWorkflowStep(ctx, workflowID, stepID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get install workflow step"))
		return
	}

	approval, err := s.getWorkflowStepApproval(ctx, org.ID, approvalID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get install workflow step approval"))
		return
	}

	ctx.JSON(http.StatusOK, approval)
}

func (s *service) getWorkflowStepApproval(ctx *gin.Context, OrgID, approvalID string) (*app.WorkflowStepApproval, error) {
	var approval app.WorkflowStepApproval
	res := s.db.WithContext(ctx).
		Where("id = ? AND org_id = ?", approvalID, OrgID).
		Preload("InstallWorkflowStep").
		Preload("Response").
		First(&approval)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get workflow step")
	}

	return &approval, nil
}
