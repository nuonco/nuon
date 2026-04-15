package service

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
)

type RetryWorkflowStepResponse struct {
	WorkflowID string `json:"workflow_id"`
	Retryable  bool   `json:"retryable"`
}

// @ID						RetryWorkflowStep
// @Summary					retry a failed or awaiting-approval workflow step
// @Description.markdown	retry_workflow_by_id.md
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
// @Success					201	{object}	RetryWorkflowStepResponse
// @Router					/v1/workflows/{workflow_id}/steps/{step_id}/retry [post]
func (s *service) RetryWorkflowStep(ctx *gin.Context) {
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
			Err: fmt.Errorf("workflow %s retry not supported for owner type", workflow.ID),
		})
		return
	}

	useQueues, err := s.featuresClient.AllFeaturesEnabled(ctx, app.OrgFeatureAppBranches, app.OrgFeatureQueues)
	if err != nil {
		ctx.Error(fmt.Errorf("checking features: %w", err))
		return
	}
	if !useQueues {
		ctx.Error(stderr.ErrUser{
			Err: fmt.Errorf("retry workflow step requires queues to be enabled"),
		})
		return
	}

	resp, err := s.flowsClient.RetryStep(ctx, &flowclient.RetryStepRequest{
		InstallWorkflowID: workflow.ID,
		StepID:            step.ID,
	})
	if err != nil {
		ctx.Error(fmt.Errorf("retry step: %w", err))
		return
	}

	ctx.JSON(201, RetryWorkflowStepResponse{
		WorkflowID: workflow.ID,
		Retryable:  resp.Retryable,
	})
}
