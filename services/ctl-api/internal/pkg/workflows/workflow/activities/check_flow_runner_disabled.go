package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

type CheckFlowRunnerDisabledRequest struct {
	FlowID string `validate:"required"`
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
		Select("id", "type", "owner_id", "owner_type").
		Where(app.Workflow{ID: req.FlowID}).
		Take(&flw)
	if res.Error != nil {
		return false, errors.Wrap(res.Error, "unable to get workflow")
	}

	if flw.OwnerType != "installs" || !flw.Type.RequiresRunner() {
		return false, nil
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
