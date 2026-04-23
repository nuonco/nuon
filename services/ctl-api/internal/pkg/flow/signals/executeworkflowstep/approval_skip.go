package executeworkflowstep

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// handleSkipResponse processes a "skip current" response.
// If the signal implements SignalWithSkipGroup and returns true, the entire
// remaining group is skipped (DirectiveSkipGroup). Otherwise only the current
// step is skipped and execution continues to the next step (DirectiveContinue).
func (s *Signal) handleSkipResponse(ctx workflow.Context, l *zap.Logger, step *app.WorkflowStep, flw *app.Workflow) error {
	l.Debug("handling approval response type: skip current and continue",
		zap.String("step_id", step.ID),
		zap.String("workflow_id", flw.ID))

	sig := stepSignal(step)

	if os, ok := sig.(signal.SignalWithOnSkip); ok {
		if err := os.OnSkip(ctx); err != nil {
			l.Warn("OnSkip hook failed", zap.Error(err))
		}
	}

	if err := s.markWorkflowApprovalPlanDenied(ctx, flw, step); err != nil {
		l.Error("failed to deny plan and update step status", zap.Error(err))
	}

	skipGroup := false
	if sg, ok := sig.(signal.SignalWithSkipGroup); ok {
		skipGroup = sg.SkipGroup()
	}

	if skipGroup {
		return writeDirective(ctx, step.ID, DirectiveSkipGroup, map[string]any{
			"step_idx": step.Idx,
			"status":   "skipped",
		})
	}

	return writeDirective(ctx, step.ID, DirectiveContinue, map[string]any{
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
