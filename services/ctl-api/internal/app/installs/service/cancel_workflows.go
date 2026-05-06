package service

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.temporal.io/api/serviceerror"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
)

type CancelWorkflowsRequest struct {
	WorkflowIDs []string `json:"workflow_ids" binding:"required"`
}

type CancelWorkflowsResponse struct {
	Cancelled []string              `json:"cancelled"`
	Errors    []CancelWorkflowError `json:"errors,omitempty"`
}

type CancelWorkflowError struct {
	WorkflowID string `json:"workflow_id"`
	Error      string `json:"error"`
}

// @ID							CancelWorkflows
// @Summary						cancel multiple workflows
// @Description					Cancel multiple workflows by ID. Returns partial results if some fail.
// @Param						body	body	CancelWorkflowsRequest	true	"workflow IDs to cancel"
// @Tags						installs
// @Accept						json
// @Produce						json
// @Security					APIKey
// @Security					OrgID
// @Failure						400	{object}	stderr.ErrResponse
// @Failure						401	{object}	stderr.ErrResponse
// @Failure						403	{object}	stderr.ErrResponse
// @Failure						500	{object}	stderr.ErrResponse
// @Success				200	{object}	CancelWorkflowsResponse
// @Router						/v1/workflows/cancel [post]
func (s *service) CancelWorkflows(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org from context: %w", err))
		return
	}

	var req CancelWorkflowsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("invalid request body: %w", err))
		return
	}

	useQueues, err := s.featuresClient.AllFeaturesEnabled(ctx, app.OrgFeatureAppBranches, app.OrgFeatureQueues)
	if err != nil {
		ctx.Error(fmt.Errorf("checking features: %w", err))
		return
	}

	resp := CancelWorkflowsResponse{}

	for _, workflowID := range req.WorkflowIDs {
		if err := s.cancelSingleWorkflow(ctx, org.ID, workflowID, useQueues); err != nil {
			resp.Errors = append(resp.Errors, CancelWorkflowError{
				WorkflowID: workflowID,
				Error:      err.Error(),
			})
		} else {
			resp.Cancelled = append(resp.Cancelled, workflowID)
		}
	}

	ctx.JSON(http.StatusOK, resp)
}

func (s *service) cancelSingleWorkflow(ctx *gin.Context, orgID, workflowID string, useQueues bool) error {
	wf, err := s.getWorkflow(ctx, orgID, workflowID)
	if err != nil {
		return fmt.Errorf("unable to get workflow: %w", err)
	}

	if !generics.SliceContains(wf.Status.Status, []app.Status{
		app.StatusInProgress,
		app.StatusPending,
		app.AwaitingApproval,
		app.Status("awaiting-approval"),
	}) {
		return fmt.Errorf("workflow is not cancelable (status: %s)", wf.Status.Status)
	}

	if err := s.cancelWorkflow(ctx, wf.ID); err != nil {
		return fmt.Errorf("unable to cancel workflow: %w", err)
	}

	if wf.Status.Status == app.StatusPending {
		return nil
	}

	if useQueues {
		step := s.findCancelableStep(wf)
		if step != nil {
			if _, err := s.flowsClient.CancelStep(ctx, &flowclient.CancelStepRequest{
				InstallWorkflowID: wf.ID,
				StepID:            step.ID,
			}); err != nil {
				s.l.Warn("failed to cancel step via queues, workflow already marked cancelled",
					zap.String("workflow_id", wf.ID),
					zap.String("step_id", step.ID),
					zap.Error(err))
			}
		}
	} else {
		id := worker.ExecuteWorkflowIDCallback(signals.RequestSignal{
			EventLoopRequest: eventloop.EventLoopRequest{
				ID: wf.OwnerID,
			},
			Signal: &signals.Signal{
				InstallWorkflowID: wf.ID,
			},
		})
		err = s.evClient.Cancel(ctx, signals.TemporalNamespace, id)
		if err != nil {
			var notFoundErr *serviceerror.NotFound
			if errors.As(err, &notFoundErr) {
				s.l.Warn("workflow canceled but not found in temporal", zap.String("workflow_id", id), zap.Error(err))
			} else {
				return fmt.Errorf("unable to cancel workflow in temporal: %w", err)
			}
		}
	}

	return nil
}
