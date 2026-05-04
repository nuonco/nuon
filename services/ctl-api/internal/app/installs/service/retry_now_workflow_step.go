package service

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
)

type RetryNowWorkflowStepResponse struct {
	WorkflowID string `json:"workflow_id"`
}

// @ID						RetryNowWorkflowStep
// @Summary					skip the auto-retry backoff for a workflow step waiting to retry
// @Description			Short-circuits the auto-retry backoff timer on a step
// @Description			currently in the waiting-to-retry state, causing the
// @Description			step to execute immediately instead of after the
// @Description			remaining backoff window.
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
// @Success					201	{object}	RetryNowWorkflowStepResponse
// @Router					/v1/workflows/{workflow_id}/steps/{step_id}/retry-now [post]
func (s *service) RetryNowWorkflowStep(ctx *gin.Context) {
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

	if step.Status.Status != app.StatusWaitingToRetry {
		ctx.Error(stderr.ErrUser{
			Err: fmt.Errorf("step is not waiting to retry (status: %s)", step.Status.Status),
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
			Err: fmt.Errorf("retry-now workflow step requires queues to be enabled"),
		})
		return
	}

	resp, err := s.flowsClient.RetryNow(ctx, &flowclient.RetryNowRequest{
		StepID: step.ID,
	})
	if err != nil {
		ctx.Error(fmt.Errorf("retry-now step: %w", err))
		return
	}

	ctx.JSON(201, RetryNowWorkflowStepResponse{
		WorkflowID: resp.WorkflowID,
	})
}
