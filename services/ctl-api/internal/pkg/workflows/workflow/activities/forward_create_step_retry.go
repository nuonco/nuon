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

type ForwardCreateStepRetryRequest struct {
	StepID string `json:"step_id" validate:"required"`
}

type ForwardCreateStepRetryResponse struct {
	StepID    string `json:"step_id"`
	NewStepID string `json:"new_step_id"`
	Directive string `json:"directive"`
}

const (
	createStepRetryBudget       = 25 * time.Second
	createStepRetryMinRemaining = time.Second
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) ForwardCreateStepRetry(ctx context.Context, req ForwardCreateStepRetryRequest) (*ForwardCreateStepRetryResponse, error) {
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

	deadline := time.Now().Add(createStepRetryBudget)
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()
	for {
		retryable, err := a.sendCreateStepRetryUpdate(ctx, &qs, req.StepID, &result)
		if err == nil {
			break
		}
		if !retryable || ctx.Err() != nil || time.Until(deadline) < createStepRetryMinRemaining {
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

func (a *Activities) sendCreateStepRetryUpdate(ctx context.Context, qs *app.QueueSignal, stepID string, result *stepRetryResult) (bool, error) {
	rawResp, err := handler.UpdateWithStart(ctx, a.tClient, qs, handler.UpdateWithStartOptions{
		UpdateName:   "create-step-retry",
		WaitForStage: tclient.WorkflowUpdateStageCompleted,
	})
	if err != nil {
		return isUnknownUpdateError(err), fmt.Errorf("unable to send create-step-retry update to step %s: %w", stepID, err)
	}
	if err := rawResp.Get(ctx, result); err != nil {
		return isUnknownUpdateError(err), updateOutcomeError(fmt.Sprintf("create-step-retry update failed for step %s", stepID), err)
	}
	return false, nil
}

func isUnknownUpdateError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "unknown update")
}
