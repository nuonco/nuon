package activities

import (
	"context"
	"fmt"

	tclient "go.temporal.io/sdk/client"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// getGroupQueueSignal fetches the step group's preloaded QueueSignal to get the
// Temporal workflow reference for sending updates.
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

// ForwardRetryStepToGroupRequest is the input for forwarding a retry-step to the group.
type ForwardRetryStepToGroupRequest struct {
	StepID      string `json:"step_id" validate:"required"`
	StepGroupID string `json:"step_group_id" validate:"required"`
}

// ForwardRetryStepToGroupResponse wraps the group's retry-step response.
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

	handle, err := a.tClient.UpdateWorkflowInNamespace(ctx, qs.Workflow.Namespace,
		tclient.UpdateWorkflowOptions{
			WorkflowID:   qs.Workflow.ID,
			UpdateName:   "retry-step",
			WaitForStage: tclient.WorkflowUpdateStageCompleted,
			Args:         []any{retryStepArg{StepID: req.StepID}},
		})
	if err != nil {
		return nil, fmt.Errorf("unable to send retry-step to group: %w", err)
	}

	var resp ForwardRetryStepToGroupResponse
	if err := handle.Get(ctx, &resp); err != nil {
		return nil, fmt.Errorf("retry-step failed on group: %w", err)
	}
	return &resp, nil
}

// ForwardCancelStepToGroupRequest is the input for forwarding a cancel-step to the group.
type ForwardCancelStepToGroupRequest struct {
	StepID      string `json:"step_id" validate:"required"`
	StepGroupID string `json:"step_group_id" validate:"required"`
}

// ForwardCancelStepToGroupResponse wraps the group's cancel-step response.
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

	_, err = a.tClient.UpdateWorkflowInNamespace(ctx, qs.Workflow.Namespace,
		tclient.UpdateWorkflowOptions{
			WorkflowID:   qs.Workflow.ID,
			UpdateName:   "cancel-step",
			WaitForStage: tclient.WorkflowUpdateStageAccepted,
			Args:         []any{cancelStepArg{StepID: req.StepID}},
		})
	if err != nil {
		return nil, fmt.Errorf("unable to send cancel-step to group: %w", err)
	}

	return &ForwardCancelStepToGroupResponse{}, nil
}

// ForwardApproveStepToGroupRequest is the input for forwarding an approve-step to the group.
type ForwardApproveStepToGroupRequest struct {
	StepID             string `json:"step_id" validate:"required"`
	StepGroupID        string `json:"step_group_id" validate:"required"`
	ApprovalResponseID string `json:"approval_response_id"`
	ResponseType       string `json:"response_type"`
}

// ForwardApproveStepToGroupResponse wraps the group's approve-step response.
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

	handle, err := a.tClient.UpdateWorkflowInNamespace(ctx, qs.Workflow.Namespace,
		tclient.UpdateWorkflowOptions{
			WorkflowID:   qs.Workflow.ID,
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
		return nil, fmt.Errorf("approve-step failed on group: %w", err)
	}
	return &resp, nil
}

// ForwardSkipStepToGroupRequest is the input for forwarding a skip-step to the group.
type ForwardSkipStepToGroupRequest struct {
	StepID      string `json:"step_id" validate:"required"`
	StepGroupID string `json:"step_group_id" validate:"required"`
}

// ForwardSkipStepToGroupResponse wraps the group's skip-step response.
type ForwardSkipStepToGroupResponse struct {
	Skippable bool `json:"skippable"`
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

	handle, err := a.tClient.UpdateWorkflowInNamespace(ctx, qs.Workflow.Namespace,
		tclient.UpdateWorkflowOptions{
			WorkflowID:   qs.Workflow.ID,
			UpdateName:   "skip-step",
			WaitForStage: tclient.WorkflowUpdateStageCompleted,
			Args:         []any{skipStepArg{StepID: req.StepID}},
		})
	if err != nil {
		return nil, fmt.Errorf("unable to send skip-step to group: %w", err)
	}

	var resp ForwardSkipStepToGroupResponse
	if err := handle.Get(ctx, &resp); err != nil {
		return nil, fmt.Errorf("skip-step failed on group: %w", err)
	}
	return &resp, nil
}
