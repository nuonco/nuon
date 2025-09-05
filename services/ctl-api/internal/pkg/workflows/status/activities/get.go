package statusactivities

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// NOTE(jm): this could probably be implemented with some type of parsing the ID to figure out what model is represented
// by it, but if we do that and something _does not_ work, then it's going to be damn near impossible to debug, so we
// keep the verbose approach here, until something more elegant comes along.
type GetStatusRequest struct {
	ID string `json:"id" validate:"required"`
}

// @temporal-gen activity
// @by-id ID
func (a *Activities) PkgStatusGetInstallWorkflowStatus(ctx context.Context, req GetStatusRequest) (*app.CompositeStatus, error) {
	var obj app.Workflow
	if err := a.getStatus(ctx, &obj, req.ID); err != nil {
		return nil, nil
	}

	return &obj.Status, nil
}

// @temporal-gen activity
// @by-id ID
func (a *Activities) PkgStatusGetInstallWorkflowStepStatus(ctx context.Context, req GetStatusRequest) (*app.CompositeStatus, error) {
	var obj app.WorkflowStep
	if err := a.getStatus(ctx, &obj, req.ID); err != nil {
		return nil, nil
	}

	return &obj.Status, nil
}

// @temporal-gen activity
// @by-id ID
func (a *Activities) PkgStatusGetInstallStackVersionStatus(ctx context.Context, req GetStatusRequest) (*app.CompositeStatus, error) {
	var obj app.InstallStackVersion
	if err := a.getStatus(ctx, &obj, req.ID); err != nil {
		return nil, nil
	}

	return &obj.Status, nil
}

func (a *Activities) getStatus(ctx context.Context, obj any, objID string) error {
	if res := a.db.WithContext(ctx).
		First(obj, "id = ?", objID); res.Error != nil {
		return errors.Wrap(res.Error, fmt.Sprintf("unable to get status for %s", objID))
	}

	return nil
}
