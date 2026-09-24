package activities

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	runnershelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

type GetStepErrorHintsRequest struct {
	StepID string `json:"step_id" validate:"required"`
}

type GetStepErrorHintsResponse struct {
	Hints compositeerrors.Hints               `json:"hints,omitempty"`
	Error *compositeerrors.CompositeErrorData `json:"error,omitempty" temporaljson:"error,omitempty"`
}

// GetStepErrorHints returns the composite error, and its hints, recorded for a
// failed step's target. The source of the error depends on the target type:
//
//   - install_stack_versions: reads the row-level CompositeError field set by
//     template render failures.
//   - install_sandbox_runs: checks the row-level CompositeError first (set by
//     plan render failures), then falls through to the latest runner job error
//     (set by infrastructure failures during apply).
//   - install_deploys: checks the row-level CompositeError first (set by plan
//     render failures), then falls through to the latest runner job error.
//   - no target (app branch run steps): reads the app branch run error for the
//     workflow, then falls back to the error already recorded on the step.
//
// It is best-effort: a target with no composite error yields empty hints.
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

	return &GetStepErrorHintsResponse{Hints: ce.Hints, Error: ce}, nil
}

// stepTargetCompositeError reads the canonical composite error for the step's
// target. Stack versions carry a row-level error set directly by the generator
// signal. Sandbox runs check the row-level error first (plan render failures),
// then fall back to the latest runner job (infrastructure failures). Deploys
// follow the same order.
func (a *Activities) stepTargetCompositeError(ctx context.Context, step *app.WorkflowStep) (*compositeerrors.CompositeErrorData, error) {
	if step.StepTargetID == "" {
		var run app.AppBranchRun
		err := a.db.WithContext(ctx).
			Select("id", "composite_error").
			Where(app.AppBranchRun{WorkflowID: &step.InstallWorkflowID}).
			First(&run).Error
		if err == nil && run.CompositeError != nil {
			return run.CompositeError, nil
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.Wrap(err, "unable to get app branch run composite error")
		}
		return step.Status.CompositeError, nil
	}

	switch app.WorkflowStepTargetType(step.StepTargetType) {
	case app.WorkflowStepTargetTypeInstallStackVersions:
		return a.stackVersionCompositeError(ctx, step.StepTargetID)

	case app.WorkflowStepTargetTypeInstallSandboxRun, app.WorkflowStepTargetTypeInstallSandboxRuns:
		rowCE, err := a.sandboxRunRowCompositeError(ctx, step.StepTargetID)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get sandbox run composite error")
		}
		if rowCE != nil {
			return rowCE, nil
		}
		jobCE, err := runnershelpers.GetLatestJobCompositeError(ctx, a.db, runnershelpers.GetLatestJobCompositeErrorRequest{
			OwnerID:   step.StepTargetID,
			OwnerType: "install_sandbox_runs",
		})
		if err != nil {
			return nil, errors.Wrap(err, "unable to get runner job composite error for sandbox run")
		}
		return jobCE, nil

	case app.WorkflowStepTargetTypeInstallDeploy, app.WorkflowStepTargetTypeInstallDeploys:
		rowCE, err := a.deployRowCompositeError(ctx, step.StepTargetID)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get install deploy composite error")
		}
		if rowCE != nil {
			return rowCE, nil
		}
		jobCE, err := runnershelpers.GetLatestJobCompositeError(ctx, a.db, runnershelpers.GetLatestJobCompositeErrorRequest{
			OwnerID:   step.StepTargetID,
			OwnerType: "install_deploys",
		})
		if err != nil {
			return nil, errors.Wrap(err, "unable to get composite error for deploy")
		}
		return jobCE, nil

	default:
		return nil, nil
	}
}

// stackVersionCompositeError reads the row-level composite error from an
// InstallStackVersion. A missing row means no row-level error, so the caller
// can fall through to the next best-effort source.
func (a *Activities) stackVersionCompositeError(ctx context.Context, stackVersionID string) (*compositeerrors.CompositeErrorData, error) {
	var sv app.InstallStackVersion
	if err := a.db.WithContext(ctx).
		Select("id", "composite_error").
		Where(app.InstallStackVersion{ID: stackVersionID}).
		First(&sv).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "unable to get stack version")
	}
	return sv.CompositeError, nil
}

func (a *Activities) deployRowCompositeError(ctx context.Context, installDeployID string) (*compositeerrors.CompositeErrorData, error) {
	var deploy app.InstallDeploy
	if err := a.db.WithContext(ctx).
		Select("id", "composite_error").
		Where(app.InstallDeploy{ID: installDeployID}).
		First(&deploy).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "unable to get install deploy")
	}
	return deploy.CompositeError, nil
}

// sandboxRunRowCompositeError reads the row-level composite error from an
// InstallSandboxRun without touching runner jobs. A missing row, or a row
// that carries no error (e.g. the run succeeded or failed via
// infrastructure), yields nil so the caller falls through to the next source.
func (a *Activities) sandboxRunRowCompositeError(ctx context.Context, sandboxRunID string) (*compositeerrors.CompositeErrorData, error) {
	var run app.InstallSandboxRun
	if err := a.db.WithContext(ctx).
		Select("id", "composite_error").
		Where(app.InstallSandboxRun{ID: sandboxRunID}).
		First(&run).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "unable to get sandbox run")
	}
	return run.CompositeError, nil
}
