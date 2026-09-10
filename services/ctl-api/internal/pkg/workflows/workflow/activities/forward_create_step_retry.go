package activities

import (
	"context"
	"fmt"
	"strings"
	"time"

	tclient "go.temporal.io/sdk/client"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
)

// ForwardCreateStepRetryRequest is the input for forwarding a create-step-retry to a step handler workflow.
type ForwardCreateStepRetryRequest struct {
	StepID string `json:"step_id" validate:"required"`
}

// ForwardCreateStepRetryResponse is the output from forwarding a create-step-retry.
type ForwardCreateStepRetryResponse struct {
	StepID    string `json:"step_id"`
	NewStepID string `json:"new_step_id"`
	Directive string `json:"directive"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) ForwardCreateStepRetry(ctx context.Context, req ForwardCreateStepRetryRequest) (*ForwardCreateStepRetryResponse, error) {
	// Find the step's handler workflow via the queue_signals table.
	var qs app.QueueSignal
	res := a.db.WithContext(ctx).
		Where(app.QueueSignal{
			OwnerID:   req.StepID,
			OwnerType: (&app.WorkflowStep{}).TableName(),
			Type:      "execute-workflow-step",
		}).
		Order("created_at DESC").
		First(&qs)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to find step queue signal for step %s: %w", req.StepID, res.Error)
	}

	var result stepRetryResult

	// Update-with-start against a completed signal starts a fresh handler run,
	// and that run accepts updates before run() registers its handlers — under
	// load the first workflow task can complete inside that window and the
	// update is rejected with "unknown update". The run registers handlers
	// moments later, so retry the send until it lands. Stay inside the 30s
	// start-to-close budget.
	deadline := time.Now().Add(20 * time.Second)
	for {
		retryable, err := a.sendCreateStepRetryUpdate(ctx, &qs, req.StepID, &result)
		if err == nil {
			break
		}
		if !retryable || time.Now().After(deadline) || ctx.Err() != nil {
			return nil, err
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("unable to send create-step-retry update to step %s: %w", req.StepID, ctx.Err())
		case <-time.After(500 * time.Millisecond):
		}
	}

	return &ForwardCreateStepRetryResponse{
		StepID:    req.StepID,
		NewStepID: result.NewStepID,
		Directive: result.Directive,
	}, nil
}

type stepRetryResult struct {
	Directive string `json:"directive"`
	NewStepID string `json:"new_step_id"`
}

// sendCreateStepRetryUpdate sends the create-step-retry update once.
// retryable reports whether the failure is the fresh-run registration window
// and is worth retrying. Update failures (e.g. max retries exhausted) stay
// concrete — wrapping them makes the activity's top-level Temporal failure
// retryable again.
func (a *Activities) sendCreateStepRetryUpdate(ctx context.Context, qs *app.QueueSignal, stepID string, result *stepRetryResult) (bool, error) {
	rawResp, err := handler.UpdateWithStart(ctx, a.tClient, qs, handler.UpdateWithStartOptions{
		UpdateName:   "create-step-retry",
		WaitForStage: tclient.WorkflowUpdateStageCompleted,
	})
	if err != nil {
		return isUnknownUpdateError(err), fmt.Errorf("unable to send create-step-retry update to step %s: %w", stepID, err)
	}
	if err := rawResp.Get(ctx, result); err != nil {
		return isUnknownUpdateError(err), err
	}
	return false, nil
}

// isUnknownUpdateError reports whether err is the worker-side rejection of an
// update delivered before the target run registered its handlers.
func isUnknownUpdateError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "unknown update")
}
