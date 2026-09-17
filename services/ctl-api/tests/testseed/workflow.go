package testseed

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// WorkflowOption allows overriding defaults when creating a Workflow.
type WorkflowOption func(*app.Workflow)

func WithWorkflowStatus(status app.CompositeStatus) WorkflowOption {
	return func(w *app.Workflow) { w.Status = status }
}

func WithWorkflowOwnerType(ownerType string) WorkflowOption {
	return func(w *app.Workflow) { w.OwnerType = ownerType }
}

func WithWorkflowPlanOnly(planOnly bool) WorkflowOption {
	return func(w *app.Workflow) { w.PlanOnly = planOnly }
}

func WithWorkflowCreatedAt(createdAt time.Time) WorkflowOption {
	return func(w *app.Workflow) { w.CreatedAt = createdAt }
}

// CreateWorkflow persists a Workflow owned by the given install.
func (s *Seeder) CreateWorkflow(ctx context.Context, t *testing.T, installID string, workflowType app.WorkflowType, opts ...WorkflowOption) *app.Workflow {
	wf := &app.Workflow{
		OwnerID:        installID,
		OwnerType:      "installs",
		Type:           workflowType,
		Status:         app.NewCompositeStatus(ctx, app.StatusPending),
		ApprovalOption: app.InstallApprovalOptionPrompt,
	}
	for _, opt := range opts {
		opt(wf)
	}
	res := s.db.WithContext(ctx).Create(wf)
	require.NoError(t, res.Error)
	return wf
}

// WorkflowStepOption allows overriding defaults when creating a WorkflowStep.
type WorkflowStepOption func(*app.WorkflowStep)

func WithStepStatus(status app.CompositeStatus) WorkflowStepOption {
	return func(s *app.WorkflowStep) { s.Status = status }
}

func WithStepRetryable(retryable bool) WorkflowStepOption {
	return func(s *app.WorkflowStep) { s.Retryable = retryable }
}

func WithStepSkippable(skippable bool) WorkflowStepOption {
	return func(s *app.WorkflowStep) { s.Skippable = skippable }
}

func WithStepSignal(signal *app.Signal) WorkflowStepOption {
	return func(s *app.WorkflowStep) { s.Signal = signal }
}

func WithStepGroup(groupID string) WorkflowStepOption {
	return func(s *app.WorkflowStep) { s.WorkflowStepGroupID = groupID }
}

// CreateWorkflowStep persists a WorkflowStep for the given workflow.
func (s *Seeder) CreateWorkflowStep(ctx context.Context, t *testing.T, workflowID string, opts ...WorkflowStepOption) *app.WorkflowStep {
	step := &app.WorkflowStep{
		InstallWorkflowID: workflowID,
		Status:            app.NewCompositeStatus(ctx, app.StatusPending),
		Signal:            &app.Signal{Type: "provision_sandbox"},
		Name:              "test-step",
		ExecutionType:     app.WorkflowStepExecutionTypeSystem,
	}
	for _, opt := range opts {
		opt(step)
	}
	tx := s.db.WithContext(ctx)
	if step.WorkflowStepGroupID == "" {
		// The column has an FK to workflow_step_groups; an empty string
		// violates it, so ungrouped steps must insert NULL.
		tx = tx.Omit("workflow_step_group_id")
	}
	res := tx.Create(step)
	require.NoError(t, res.Error)
	return step
}

// CreateWorkflowStepGroup persists a WorkflowStepGroup for the given workflow.
func (s *Seeder) CreateWorkflowStepGroup(ctx context.Context, t *testing.T, workflowID string) *app.WorkflowStepGroup {
	group := &app.WorkflowStepGroup{
		WorkflowID: workflowID,
		Name:       "test-group",
		Status:     app.NewCompositeStatus(ctx, app.StatusPending),
	}
	res := s.db.WithContext(ctx).Create(group)
	require.NoError(t, res.Error)
	return group
}

// CreateWorkflowStepApproval persists a WorkflowStepApproval for the given step.
func (s *Seeder) CreateWorkflowStepApproval(ctx context.Context, t *testing.T, stepID string, approvalType app.WorkflowStepApprovalType, contents string) *app.WorkflowStepApproval {
	approval := &app.WorkflowStepApproval{
		InstallWorkflowStepID: stepID,
		Type:                  approvalType,
		Contents:              contents,
	}
	res := s.db.WithContext(ctx).Create(approval)
	require.NoError(t, res.Error)
	return approval
}
