package activities

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

type GetStepErrorHintsRequest struct {
	StepID string `json:"step_id" validate:"required"`
}

type GetStepErrorHintsResponse struct {
	Hints compositeerrors.Hints `json:"hints,omitempty"`
}

// GetStepErrorHints returns the composite-error hints recorded for a failed
// step's target. The runner-result chokepoint parses a failed execution into a
// typed composite error and stores it (with its hints) on the execution's
// result row. This activity reads those hints back so the step orchestrator can
// act on them (e.g. skip auto-retries for a failure that won't resolve by
// retrying).
//
// It is best-effort: a target with no composite error, or a target type that
// carries no runner job, yields empty hints.
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

// stepTargetCompositeError reads the composite error off the step target's
// latest runner job execution result. This is the canonical, per-attempt copy:
// the runner posts the result before the job is marked terminal, so the row is
// visible by the time the step wakes and this activity runs. Reading it here
// rather than the mirrored install_deploys / install_sandbox_runs aggregate
// column avoids racing the chokepoint's best-effort aggregate write.
//
// Only deploy/sandbox-run targets own runner jobs with parsed composite errors;
// any other target (or an unset target) yields nil. A missing runner
// job/execution/result means no hint was recorded, so we yield nil rather than
// failing.
func (a *Activities) stepTargetCompositeError(ctx context.Context, step *app.WorkflowStep) (*compositeerrors.CompositeErrorData, error) {
	if step.StepTargetID == "" {
		return nil, nil
	}

	switch app.WorkflowStepTargetType(step.StepTargetType) {
	case app.WorkflowStepTargetTypeInstallDeploy, app.WorkflowStepTargetTypeInstallDeploys,
		app.WorkflowStepTargetTypeInstallSandboxRun, app.WorkflowStepTargetTypeInstallSandboxRuns:
	default:
		return nil, nil
	}

	runnerJob, err := a.getRunnerJob(ctx, &GetRunnerJobRequest{RunnerJobOwnerID: step.StepTargetID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "unable to get runner job for step target")
	}

	execution, err := a.getRunnerJobExecution(ctx, GetRunnerJobExecutionRequest{RunnerJobID: runnerJob.ID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "unable to get runner job execution for step target")
	}

	result, err := a.getRunnerJobExecutionResult(ctx, GetRunnerJobExecutionResultRequest{RunnerJobExecutionID: execution.ID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "unable to get runner job execution result for step target")
	}

	return result.CompositeError, nil
}
