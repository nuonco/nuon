package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
)

type SkipWorkflowStepResponse struct {
	WorkflowID string `json:"workflow_id"`
	Skippable  bool   `json:"skippable"`
}

// @ID						SkipWorkflowStep
// @Summary					skip a failed workflow step and continue the workflow
// @Description.markdown	skip_workflow_step.md
// @Param					workflow_id	path	string	true	"workflow ID"
// @Param					step_id		path	string	true	"step ID"
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
// @Success					201	{object}	SkipWorkflowStepResponse
// @Router					/v1/workflows/{workflow_id}/steps/{step_id}/skip [post]
func (s *service) SkipWorkflowStep(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(stderr.ErrUser{
			Err: fmt.Errorf("unable to get org from context: %w", err),
		})
		return
	}

	workflowID := ctx.Param("workflow_id")
	workflow, err := s.getWorkflow(ctx, org.ID, workflowID)
	if err != nil {
		ctx.Error(stderr.ErrUser{
			Err: fmt.Errorf("workflow not found: %w", err),
		})
		return
	}

	stepID := ctx.Param("step_id")
	step, err := s.getWorkflowStep(ctx, org.ID, workflow.ID, stepID)
	if err != nil {
		ctx.Error(stderr.ErrUser{
			Err: fmt.Errorf("workflow step not found: %w", err),
		})
		return
	}

	if workflow.OwnerType != "installs" {
		ctx.Error(stderr.ErrUser{
			Err: fmt.Errorf("workflow %s skip not supported for owner type", workflow.ID),
		})
		return
	}

	resp, err := s.flowsClient.SkipStep(ctx, &flowclient.SkipStepRequest{
		InstallWorkflowID: workflow.ID,
		StepID:            step.ID,
	})
	if err != nil {
		ctx.Error(fmt.Errorf("skip step: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, SkipWorkflowStepResponse{
		WorkflowID: workflow.ID,
		Skippable:  resp.Skippable,
	})
}
