package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type GetStepErrorHintsRequest struct {
	StepID string `json:"step_id" validate:"required"`
}

type GetStepErrorHintsResponse struct {
	Hints compositeerrors.Hints `json:"hints,omitempty"`
}

// GetStepErrorHints returns the composite-error hints recorded for a failed
// step's target. The runner-result chokepoint parses a failed execution into a
// typed composite error and mirrors it (with its hints) onto the target
// aggregate row (install_deploys / install_sandbox_runs). This activity reads
// those hints back so the step orchestrator can act on them (e.g. skip
// auto-retries for a failure that won't resolve by retrying).
//
// It is best-effort: a target with no composite error, or a target type that
// carries no composite_error column, yields empty hints.
//
// @temporal-gen-v2 activity
// @max-retries 1
func (a *Activities) GetStepErrorHints(ctx context.Context, req GetStepErrorHintsRequest) (*GetStepErrorHintsResponse, error) {
	step, err := a.PkgWorkflowsFlowGetFlowsStep(ctx, GetFlowStepRequest{
		FlowStepID: req.StepID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get step")
	}

	ce, err := a.stepTargetCompositeError(ctx, step)
	if err != nil {
		return nil, err
	}
	if ce == nil {
		return &GetStepErrorHintsResponse{}, nil
	}

	return &GetStepErrorHintsResponse{Hints: ce.Hints}, nil
}

// stepTargetCompositeError reads the composite_error column off the step's
// target aggregate row. Only target types that carry the column are handled;
// any other target (or an unset target) yields nil.
func (a *Activities) stepTargetCompositeError(ctx context.Context, step *app.WorkflowStep) (*compositeerrors.CompositeErrorData, error) {
	if step.StepTargetID == "" {
		return nil, nil
	}

	switch app.WorkflowStepTargetType(step.StepTargetType) {
	case app.WorkflowStepTargetTypeInstallDeploy, app.WorkflowStepTargetTypeInstallDeploys:
		var deploy app.InstallDeploy
		if err := a.db.WithContext(ctx).
			Select("composite_error").
			First(&deploy, "id = ?", step.StepTargetID).Error; err != nil {
			// A missing target row is terminal (won't resolve by retrying); a
			// transient DB error stays retryable.
			return nil, generics.TemporalGormError(err, "unable to get install deploy")
		}
		return deploy.CompositeError, nil
	case app.WorkflowStepTargetTypeInstallSandboxRun, app.WorkflowStepTargetTypeInstallSandboxRuns:
		var run app.InstallSandboxRun
		if err := a.db.WithContext(ctx).
			Select("composite_error").
			First(&run, "id = ?", step.StepTargetID).Error; err != nil {
			return nil, generics.TemporalGormError(err, "unable to get install sandbox run")
		}
		return run.CompositeError, nil
	default:
		return nil, nil
	}
}
