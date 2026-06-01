package activities

import (
	"context"
	"fmt"

	tclient "go.temporal.io/sdk/client"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

type ApprovePlanRequest struct {
	InstallWorkflowID  string                       `validate:"required"`
	StepID             string                       `validate:"required"`
	ApprovalResponseID string                       `validate:"required"`
	ResponseType       app.WorkflowStepResponseType `validate:"required"`
}

type ApprovePlanResponse struct{}

// approveStepArgs mirrors executeflow.ApproveStepRequest. It is duplicated here
// (rather than imported) to avoid an import cycle:
//
//	installs/worker/activities -> flow/signals/executeflow -> installs/signals/...
//	  -> installs/worker/activities
//
// The wire format must stay in sync with executeflow.ApproveStepRequest.
type approveStepArgs struct {
	StepID             string `json:"step_id"`
	ApprovalResponseID string `json:"approval_response_id"`
	ResponseType       string `json:"response_type"`
}

// ApprovePlan forwards an approval response to the running install workflow by
// sending the "approve-step" Temporal update to the execute-flow handler workflow.
//
// This is the same operation the API used to perform synchronously via
// flow/client.ApprovePlan; calling it from a Nuon Signal's Execute() phase gives
// us first-class lifecycle webhooks and queue-backed retries while keeping the
// underlying parent-workflow wakeup mechanism unchanged.
//
// @temporal-gen-v2 activity
func (a *Activities) ApprovePlan(ctx context.Context, req *ApprovePlanRequest) (*ApprovePlanResponse, error) {
	var qs app.QueueSignal
	res := a.db.WithContext(ctx).
		Where(app.QueueSignal{
			OwnerID:   req.InstallWorkflowID,
			OwnerType: "install_workflows",
			// "execute-workflow" is executeflow.SignalType; duplicated as a literal
			// to avoid the import cycle described on approveStepArgs.
			Type: signal.SignalType("execute-workflow"),
		}).
		Order("created_at DESC").
		First(&qs)
	if res.Error != nil {
		return nil, fmt.Errorf("queue signal not found for install_workflow %s: %w", req.InstallWorkflowID, res.Error)
	}

	handle, err := handler.UpdateWithStart(ctx, a.tClient, &qs, handler.UpdateWithStartOptions{
		UpdateName:   "approve-step",
		WaitForStage: tclient.WorkflowUpdateStageCompleted,
		Args: []any{
			approveStepArgs{
				StepID:             req.StepID,
				ApprovalResponseID: req.ApprovalResponseID,
				ResponseType:       string(req.ResponseType),
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("unable to send approve-step update: %w", err)
	}

	// Drain the response (we don't use it, but Get blocks until the update is
	// applied, surfacing any in-workflow validation errors).
	var resp struct{}
	if err := handle.Get(ctx, &resp); err != nil {
		return nil, fmt.Errorf("approve-step update failed: %w", err)
	}

	return &ApprovePlanResponse{}, nil
}
