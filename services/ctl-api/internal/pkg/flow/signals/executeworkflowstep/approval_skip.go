package executeworkflowstep

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// handleSkipResponse processes a "skip current" response.
func (s *Signal) handleSkipResponse(ctx workflow.Context, l *zap.Logger, step *app.WorkflowStep, flw *app.Workflow) error {
	l.Debug("handling approval response type: skip current and continue",
		zap.String("step_id", step.ID),
		zap.String("workflow_id", flw.ID))

	if os, ok := stepSignal(step).(signal.SignalWithOnSkip); ok {
		if err := os.OnSkip(ctx); err != nil {
			l.Warn("OnSkip hook failed", zap.Error(err))
		}
	}

	if err := s.markWorkflowApprovalPlanDenied(ctx, flw, step); err != nil {
		l.Error("failed to deny plan and update step status", zap.Error(err))
	}

	return writeDirective(ctx, step.ID, DirectiveSkipGroup, map[string]any{
		"step_idx": step.Idx,
		"status":   "skipped",
	})
}

// handleSkipDependentsResponse processes a "skip current and dependents" response.
func (s *Signal) handleSkipDependentsResponse(ctx workflow.Context, l *zap.Logger, step *app.WorkflowStep, flw *app.Workflow) error {
	l.Debug("handling approval response type: skip current and dependents",
		zap.String("step_id", step.ID),
		zap.String("workflow_id", flw.ID))

	if os, ok := stepSignal(step).(signal.SignalWithOnSkip); ok {
		if err := os.OnSkip(ctx); err != nil {
			l.Warn("OnSkip hook failed", zap.Error(err))
		}
	}

	if err := s.markDependentStepsAsSkipped(ctx, flw, step); err != nil {
		l.Error("failed to deny plan and update step status", zap.Error(err))
	}

	return writeDirective(ctx, step.ID, DirectiveStop, map[string]any{
		"step_idx": step.Idx,
		"status":   "stopped",
	})
}
