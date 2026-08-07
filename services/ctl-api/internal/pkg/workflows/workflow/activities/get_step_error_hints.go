package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	runnershelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
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

// stepTargetCompositeError reads the canonical composite error from the step
// target's latest runner job execution result. A missing job, execution, or
// result means no hint was recorded and yields nil.
func (a *Activities) stepTargetCompositeError(ctx context.Context, step *app.WorkflowStep) (*compositeerrors.CompositeErrorData, error) {
	if step.StepTargetID == "" {
		return nil, nil
	}

	var ownerType string
	switch app.WorkflowStepTargetType(step.StepTargetType) {
	case app.WorkflowStepTargetTypeInstallDeploy, app.WorkflowStepTargetTypeInstallDeploys:
		ownerType = "install_deploys"
	case app.WorkflowStepTargetTypeInstallSandboxRun, app.WorkflowStepTargetTypeInstallSandboxRuns:
		ownerType = "install_sandbox_runs"
	case app.WorkflowStepTargetTypeInstallActionWorkflowRun, app.WorkflowStepTargetTypeInstallActionWorkflowRuns:
		ownerType = "install_action_workflow_runs"
	default:
		return nil, nil
	}

	compositeError, err := runnershelpers.GetLatestJobCompositeError(ctx, a.db, runnershelpers.GetLatestJobCompositeErrorRequest{
		OwnerID:   step.StepTargetID,
		OwnerType: ownerType,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get composite error for step target")
	}

	return compositeError, nil
}
