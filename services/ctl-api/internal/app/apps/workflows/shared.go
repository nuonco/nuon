package workflows

import (
	"github.com/jackc/pgx/v5/pgtype"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

// WorkflowStepOptions is a functional option for configuring WorkflowStep
type WorkflowStepOptions func(*app.WorkflowStep)

// WithSkippable sets whether a step can be skipped
func WithSkippable(skippable bool) WorkflowStepOptions {
	return func(s *app.WorkflowStep) {
		s.Skippable = skippable
	}
}

// WithGroupIdx sets the group index for a step
func WithGroupIdx(n int) WorkflowStepOptions {
	return func(s *app.WorkflowStep) {
		s.GroupIdx = n
	}
}

// stepGroup tracks the current group index for workflow steps
type stepGroup struct {
	idx int
}

// newStepGroup creates a new step group starting at index 0
func newStepGroup() *stepGroup {
	return &stepGroup{}
}

// nextGroup increments the group index for sequential execution
func (s *stepGroup) nextGroup() {
	s.idx++
}

// appBranchSignalStep creates a WorkflowStep from an app branch signal
func (s *stepGroup) appBranchSignalStep(ctx workflow.Context, appBranchID, name string, metadata pgtype.Hstore, sig signal.Signal, opts ...WorkflowStepOptions) (*app.WorkflowStep, error) {
	opts = append(opts, WithGroupIdx(s.idx))
	return appBranchSignalStep(ctx, appBranchID, name, metadata, sig, opts...)
}

// appBranchSignalStep creates a workflow step from a signal
func appBranchSignalStep(ctx workflow.Context, appBranchID, name string, metadata pgtype.Hstore, sig signal.Signal, opts ...WorkflowStepOptions) (*app.WorkflowStep, error) {
	if sig == nil {
		step := &app.WorkflowStep{
			Name:          name,
			ExecutionType: app.WorkflowStepExecutionTypeSkipped,
			Status:        app.NewCompositeTemporalStatus(ctx, app.StatusPending),
			Metadata:      metadata,
		}

		for _, o := range opts {
			o(step)
		}

		return step, nil
	}

	// Default execution type is system
	executionTyp := app.WorkflowStepExecutionTypeSystem

	// Set target type based on signal type
	// For now, use app_branches as the default target type for all app branch signals
	targetType := "app_branches"

	step := &app.WorkflowStep{
		Name:           name,
		ExecutionType:  executionTyp,
		StepTargetType: targetType,
		OwnerID:        appBranchID,
		OwnerType:      "app_branches",
		Status:         app.NewCompositeTemporalStatus(ctx, app.StatusPending),
		Metadata:       metadata,
		QueueSignal: &signaldb.SignalData{
			Signal: sig,
		},
		Retryable: true,
		Skippable: true,
	}

	for _, o := range opts {
		o(step)
	}

	return step, nil
}
