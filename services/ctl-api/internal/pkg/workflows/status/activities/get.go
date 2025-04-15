package statusactivities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// NOTE(jm): this could probably be implemented with some type of parsing the ID to figure out what model is represented
// by it, but if we do that and something _does not_ work, then it's going to be damn near impossible to debug, so we
// keep the verbose approach here, until something more elegant comes along.
type GetStatusRequest struct {
	ID string `json:"id"`
}

// @temporal-gen activity
// @by-id ID
func (a *Activities) PkgStatusGetInstallWorkflowStatus(ctx context.Context, req GetStatusRequest) (*app.CompositeStatus, error) {
	var obj app.InstallWorkflow
	if err := a.getStatus(ctx, &obj, req.ID); err != nil {
		return nil, nil
	}

	return &obj.Status, nil
}

// @temporal-gen activity
// @by-id ID
func (a *Activities) PkgStatusGetInstallWorkflowStepStatus(ctx context.Context, req GetStatusRequest) (*app.CompositeStatus, error) {
	var obj app.InstallWorkflowStep
	if err := a.getStatus(ctx, &obj, req.ID); err != nil {
		return nil, nil
	}

	return &obj.Status, nil
}

// @temporal-gen activity
// @by-id ID
func (a *Activities) PkgStatusGetInstallCloudFormationStackVersionStatus(ctx context.Context, req GetStatusRequest) (*app.CompositeStatus, error) {
	var obj app.InstallAWSCloudFormationStackVersion
	if err := a.getStatus(ctx, &obj, req.ID); err != nil {
		return nil, nil
	}

	return &obj.Status, nil
}

// @temporal-gen activity
// @by-id ID
func (a *Activities) PkgStatusGetInstallCloudFormationStackRunStatus(ctx context.Context, req GetStatusRequest) (*app.CompositeStatus, error) {
	var obj app.InstallAWSCloudFormationStackRun
	if err := a.getStatus(ctx, &obj, req.ID); err != nil {
		return nil, nil
	}

	return &obj.Status, nil
}

func (a *Activities) getStatus(ctx context.Context, obj any, objID string) error {
	if res := a.db.WithContext(ctx).
		First(obj, "id = ?", objID); res.Error != nil {
		return errors.Wrap(res.Error, "unable to get status")
	}

	return nil
}
