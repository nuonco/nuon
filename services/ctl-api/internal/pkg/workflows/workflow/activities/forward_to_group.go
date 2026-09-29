package activities

import (
	"context"
	"fmt"

	tclient "go.temporal.io/sdk/client"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
)

func (a *Activities) getGroupQueueSignal(ctx context.Context, stepGroupID string) (*app.QueueSignal, error) {
	var group app.WorkflowStepGroup
	res := a.db.WithContext(ctx).
		Preload("QueueSignal").
		Where(app.WorkflowStepGroup{ID: stepGroupID}).
		First(&group)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to find step group %s: %w", stepGroupID, res.Error)
	}
	if group.QueueSignal == nil {
		return nil, fmt.Errorf("step group %s has no queue signal", stepGroupID)
	}
	return group.QueueSignal, nil
}

type ForwardRetryStepToGroupRequest struct {
	StepID      string `json:"step_id" validate:"required"`
	StepGroupID string `json:"step_group_id" validate:"required"`
}

type ForwardRetryStepToGroupResponse struct {
	Retryable bool `json:"retryable"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) ForwardRetryStepToGroup(ctx context.Context, req ForwardRetryStepToGroupRequest) (*ForwardRetryStepToGroupResponse, error) {
	qs, err := a.getGroupQueueSignal(ctx, req.StepGroupID)
	if err != nil {
		return nil, err
	}

	type retryStepArg struct {
		StepID string `json:"step_id"`
	}

	handle, err := handler.UpdateWithStart(ctx, a.tClient, qs, handler.UpdateWithStartOptions{
		UpdateName:   "retry-step",
		WaitForStage: tclient.WorkflowUpdateStageCompleted,
		Args:         []any{retryStepArg{StepID: req.StepID}},
	})
	if err != nil {
		return nil, fmt.Errorf("unable to send retry-step to group: %w", err)
	}

	var resp ForwardRetryStepToGroupResponse
	if err := handle.Get(ctx, &resp); err != nil {
		return nil, updateOutcomeError("retry-step failed on group", err)
	}
	return &resp, nil
}

type ForwardCancelStepToGroupRequest struct {
	StepID      string `json:"step_id" validate:"required"`
	StepGroupID string `json:"step_group_id" validate:"required"`
}

type ForwardCancelStepToGroupResponse struct{}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) ForwardCancelStepToGroup(ctx context.Context, req ForwardCancelStepToGroupRequest) (*ForwardCancelStepToGroupResponse, error) {
	qs, err := a.getGroupQueueSignal(ctx, req.StepGroupID)
	if err != nil {
		return nil, err
	}

	type cancelStepArg struct {
		StepID string `json:"step_id"`
	}

	_, err = handler.UpdateWithStart(ctx, a.tClient, qs, handler.UpdateWithStartOptions{
		UpdateName:   "cancel-step",
		WaitForStage: tclient.WorkflowUpdateStageAccepted,
		Args:         []any{cancelStepArg{StepID: req.StepID}},
	})
	if err != nil {
		return nil, fmt.Errorf("unable to send cancel-step to group: %w", err)
	}

	return &ForwardCancelStepToGroupResponse{}, nil
}

type ForwardApproveStepToGroupRequest struct {
	StepID             string `json:"step_id" validate:"required"`
	StepGroupID        string `json:"step_group_id" validate:"required"`
	ApprovalResponseID string `json:"approval_response_id"`
	ResponseType       string `json:"response_type"`
}

type ForwardApproveStepToGroupResponse struct{}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) ForwardApproveStepToGroup(ctx context.Context, req ForwardApproveStepToGroupRequest) (*ForwardApproveStepToGroupResponse, error) {
	qs, err := a.getGroupQueueSignal(ctx, req.StepGroupID)
	if err != nil {
		return nil, err
	}

	type approveStepArg struct {
		StepID             string `json:"step_id"`
		ApprovalResponseID string `json:"approval_response_id"`
		ResponseType       string `json:"response_type"`
	}

	handle, err := handler.UpdateWithStart(ctx, a.tClient, qs, handler.UpdateWithStartOptions{
		UpdateName:   "approve-step",
		WaitForStage: tclient.WorkflowUpdateStageCompleted,
		Args: []any{approveStepArg{
			StepID:             req.StepID,
			ApprovalResponseID: req.ApprovalResponseID,
			ResponseType:       req.ResponseType,
		}},
	})
	if err != nil {
		return nil, fmt.Errorf("unable to send approve-step to group: %w", err)
	}

	var resp ForwardApproveStepToGroupResponse
	if err := handle.Get(ctx, &resp); err != nil {
		return nil, updateOutcomeError("approve-step failed on group", err)
	}
	return &resp, nil
}

type ForwardSkipStepToGroupRequest struct {
	StepID      string `json:"step_id" validate:"required"`
	StepGroupID string `json:"step_group_id" validate:"required"`
}

type ForwardSkipStepToGroupResponse struct {
	Skippable bool   `json:"skippable"`
	Directive string `json:"directive,omitempty"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) ForwardSkipStepToGroup(ctx context.Context, req ForwardSkipStepToGroupRequest) (*ForwardSkipStepToGroupResponse, error) {
	qs, err := a.getGroupQueueSignal(ctx, req.StepGroupID)
	if err != nil {
		return nil, err
	}

	type skipStepArg struct {
		StepID string `json:"step_id"`
	}

	handle, err := handler.UpdateWithStart(ctx, a.tClient, qs, handler.UpdateWithStartOptions{
		UpdateName:   "skip-step",
		WaitForStage: tclient.WorkflowUpdateStageCompleted,
		Args:         []any{skipStepArg{StepID: req.StepID}},
	})
	if err != nil {
		return nil, fmt.Errorf("unable to send skip-step to group: %w", err)
	}

	var resp ForwardSkipStepToGroupResponse
	if err := handle.Get(ctx, &resp); err != nil {
		return nil, updateOutcomeError("skip-step failed on group", err)
	}
	return &resp, nil
}
