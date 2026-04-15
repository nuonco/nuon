package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
)

type CancelWorkflowStepResponse struct {
	WorkflowID string `json:"workflow_id"`
}

// @ID						CancelWorkflowStep
// @Summary					cancel an in-progress workflow step
// @Description.markdown	cancel_workflow_step.md
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
// @Success					202	{object}	CancelWorkflowStepResponse
// @Router					/v1/workflows/{workflow_id}/steps/{step_id}/cancel [post]
func (s *service) CancelWorkflowStep(ctx *gin.Context) {
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

	cancelableStatuses := []app.Status{
		app.StatusInProgress,
		app.StatusPending,
		app.AwaitingApproval,
		app.Status("awaiting-approval"),
	}
	cancelable := false
	for _, status := range cancelableStatuses {
		if step.Status.Status == status {
			cancelable = true
			break
		}
	}
	if !cancelable {
		ctx.Error(stderr.ErrUser{
			Err: fmt.Errorf("workflow step %s is not cancelable (status: %s)", step.ID, step.Status.Status),
		})
		return
	}

	useQueues, err := s.featuresClient.AllFeaturesEnabled(ctx, app.OrgFeatureAppBranches, app.OrgFeatureQueues)
	if err != nil {
		ctx.Error(fmt.Errorf("checking features: %w", err))
		return
	}

	if useQueues {
		if _, err := s.flowsClient.CancelStep(ctx, &flowclient.CancelStepRequest{
			InstallWorkflowID: workflow.ID,
			StepID:            step.ID,
		}); err != nil {
			ctx.Error(fmt.Errorf("cancel step: %w", err))
			return
		}
	} else {
		ctx.Error(stderr.ErrUser{
			Err: fmt.Errorf("cancel step is only supported for queue-based workflows"),
		})
		return
	}

	ctx.JSON(http.StatusAccepted, CancelWorkflowStepResponse{
		WorkflowID: workflow.ID,
	})
}
