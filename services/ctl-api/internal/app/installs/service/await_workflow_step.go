package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

const awaitWorkflowStepTimeout = 60 * time.Second

// @ID						AwaitWorkflowStep
// @Summary					long-poll for workflow step completion
// @Description.markdown	await_workflow_step.md
// @Param					workflow_id	path	string	true	"workflow id"
// @Param					step_id		path	string	true	"step id"
// @Tags					installs
// @Accept					json
// @Produce					json
// @Security				APIKey
// @Security				OrgID
// @Failure					400	{object}	stderr.ErrResponse
// @Failure					401	{object}	stderr.ErrResponse
// @Failure					404	{object}	stderr.ErrResponse
// @Failure					408	{object}	stderr.ErrResponse
// @Failure					500	{object}	stderr.ErrResponse
// @Success					200	{object}	app.WorkflowStep
// @Router					/v1/workflows/{workflow_id}/steps/{step_id}/await [GET]
func (s *service) AwaitWorkflowStep(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org from context: %w", err))
		return
	}

	workflowID := ctx.Param("workflow_id")
	stepID := ctx.Param("step_id")

	step, err := s.getWorkflowStep(ctx, org.ID, workflowID, stepID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get workflow step: %w", err))
		return
	}

	var queueSignal app.QueueSignal
	if res := s.db.WithContext(ctx).
		Where("owner_id = ?", step.ID).
		Where("owner_type = ?", "install_workflow_steps").
		First(&queueSignal); res.Error != nil {
		ctx.Error(fmt.Errorf("unable to find queue signal for step %s: %w", step.ID, res.Error))
		return
	}

	awaitCtx, cancel := context.WithTimeout(ctx.Request.Context(), awaitWorkflowStepTimeout)
	defer cancel()

	_, err = s.queueClient.AwaitSignal(awaitCtx, queueSignal.ID)
	if err != nil {
		if awaitCtx.Err() != nil {
			ctx.JSON(http.StatusRequestTimeout, gin.H{"error": "timeout waiting for workflow step completion"})
			return
		}
		ctx.Error(fmt.Errorf("error awaiting workflow step: %w", err))
		return
	}

	finalStep, err := s.getWorkflowStep(ctx, org.ID, workflowID, stepID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get workflow step: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, finalStep)
}
