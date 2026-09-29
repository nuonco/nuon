package statusactivities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type GetStatusRequest struct {
	ID string `json:"id" validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field ID
func (a *Activities) PkgStatusGetInstallWorkflowStatus(ctx context.Context, req GetStatusRequest) (*app.CompositeStatus, error) {
	var obj app.Workflow
	if err := a.getStatus(ctx, &obj, req.ID); err != nil {
		return nil, nil
	}

	return &obj.Status, nil
}

// @temporal-gen-v2 activity
// @by-field ID
func (a *Activities) PkgStatusGetInstallWorkflowStepStatus(ctx context.Context, req GetStatusRequest) (*app.CompositeStatus, error) {
	var obj app.WorkflowStep
	if err := a.getStatus(ctx, &obj, req.ID); err != nil {
		return nil, nil
	}

	return &obj.Status, nil
}

// @temporal-gen-v2 activity
// @by-field ID
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
		return generics.TemporalGormError(res.Error, fmt.Sprintf("unable to get status for %s", objID))
	}

	return nil
}
