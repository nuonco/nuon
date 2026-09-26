package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

type CheckFlowRunnerDisabledRequest struct {
	FlowID string `validate:"required"`

	// HonorStackChanged defers the disabled-runner stop until the stack apply
	// step has finished. Empty on histories that started before that deferral.
	HonorStackChanged bool `temporaljson:"honor_stack_changed,omitempty"`
}

// @temporal-gen-v2 activity
// @max-retries 3
//
// CheckFlowRunnerDisabled reports whether an install-owned workflow needs the
// install runner but the stack has it disabled. Workflow creation already
// rejects this, so a true result means the runner was disabled after the
// workflow started.
func (a *Activities) CheckFlowRunnerDisabled(ctx context.Context, req CheckFlowRunnerDisabledRequest) (bool, error) {
	var flw app.Workflow
	res := a.db.WithContext(ctx).
		Scopes(scopes.WithDisableViews).
		Select("id", "type", "owner_id", "owner_type", "metadata").
		Where(app.Workflow{ID: req.FlowID}).
		Take(&flw)
	if res.Error != nil {
		return false, errors.Wrap(res.Error, "unable to get workflow")
	}

	if flw.OwnerType != "installs" || !flw.Type.RequiresInstallRunner() {
		return false, nil
	}

	if req.HonorStackChanged && flw.IsStackChanged() {
		finished, err := a.stackAwaitFinished(ctx, flw.ID)
		if err != nil {
			return false, err
		}
		if !finished {
			return false, nil
		}
	}

	groupIDs := a.db.WithContext(ctx).
		Model(&app.RunnerGroup{}).
		Select("id").
		Where(app.RunnerGroup{OwnerID: flw.OwnerID, OwnerType: "installs"})

	var statuses []app.RunnerStatus
	res = a.db.WithContext(ctx).
		Model(&app.Runner{}).
		Scopes(scopes.WithDisableViews).
		Where("runner_group_id IN (?)", groupIDs).
		Order("created_at DESC").
		Limit(1).
		Pluck("status", &statuses)
	if res.Error != nil {
		return false, errors.Wrap(res.Error, "unable to get install runner status")
	}
	if len(statuses) == 0 {
		return false, nil
	}

	return statuses[0] == app.RunnerStatusDisabled, nil
}

// stackAwaitFinished reports whether the stack apply and the following runner
// startup wait have both finished. The startup wait is what lets a replacement
// runner come up; stopping before it would reject the workflow the stack just
// repaired. A missing step means generation has not persisted it yet.
func (a *Activities) stackAwaitFinished(ctx context.Context, flowID string) (bool, error) {
	var steps []app.WorkflowStep
	res := a.db.WithContext(ctx).
		Select("name", "status").
		Where(app.WorkflowStep{InstallWorkflowID: flowID}).
		Find(&steps)
	if res.Error != nil {
		return false, errors.Wrap(res.Error, "unable to get stack steps")
	}

	awaitDone := false
	runnerWaitDone := false
	for _, step := range steps {
		switch step.Name {
		case app.AwaitInstallStackStepName:
			if !stackStepTerminal(step.Status.Status) {
				return false, nil
			}
			awaitDone = true
		case app.RunnerHealthyStepName:
			if !stackStepTerminal(step.Status.Status) {
				return false, nil
			}
			runnerWaitDone = true
		}
	}
	return awaitDone && runnerWaitDone, nil
}

func stackStepTerminal(status app.Status) bool {
	switch status {
	case app.StatusSuccess, app.StatusAutoSkipped, app.StatusUserSkipped,
		app.StatusDiscarded, app.StatusCancelled, app.StatusError,
		app.StatusNotAttempted,
		app.WorkflowStepApprovalStatusApproved, app.WorkflowStepApprovalStatusApprovalDenied,
		app.WorkflowStepApprovalStatusApprovalExpired,
		app.WorkflowStepNoDrift, app.WorkflowStepDrifted:
		return true
	default:
		return false
	}
}
