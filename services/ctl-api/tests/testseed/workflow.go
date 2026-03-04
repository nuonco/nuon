package testseed

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// CreateWorkflow persists a Workflow owned by the given install.
func (s *Seeder) CreateWorkflow(ctx context.Context, t *testing.T, installID string, workflowType app.WorkflowType) *app.Workflow {
	wf := &app.Workflow{
		OwnerID:   installID,
		OwnerType: "installs",
		Type:      workflowType,
		Status:    app.NewCompositeStatus(ctx, app.StatusPending),
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
	res := s.db.WithContext(ctx).Create(step)
	require.NoError(t, res.Error)
	return step
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
