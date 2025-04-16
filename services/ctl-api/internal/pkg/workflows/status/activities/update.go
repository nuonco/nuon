package statusactivities

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

// NOTE
//
// This package is the beginning of consolidating all status logic into a single package.
//
// Right now, it's a bit verbose with getters for statuses when updating, however long term we can either generate this
// or make the status selectable in isolation by selecting the field using reflection or something else.
//
// However, for now, this interface provides a few things:
// 1. ability to manage history of a status
// 2. ability to start doing things such as sending a signal to a channel if needed. This enables the ability to start
// blocking for a "status" change or a specific status.
type UpdateStatusRequest struct {
	ID     string
	Status app.CompositeStatus `json:"status"`
}

// @temporal-gen activity
func (a *Activities) PkgStatusUpdateInstallWorkflowStatus(ctx context.Context, req UpdateStatusRequest) error {
	obj := app.InstallWorkflow{
		ID: req.ID,
	}

	getter := func(ctx context.Context) (app.CompositeStatus, error) {
		var obj app.InstallWorkflow
		if err := a.getStatus(ctx, &obj, req.ID); err != nil {
			return app.CompositeStatus{}, err
		}

		return obj.Status, nil
	}

	return a.updateStatus(ctx, &obj, req.Status, getter)
}

// @temporal-gen activity
func (a *Activities) PkgStatusUpdateInstallWorkflowStepStatus(ctx context.Context, req UpdateStatusRequest) error {
	obj := app.InstallWorkflowStep{
		ID: req.ID,
	}

	getter := func(ctx context.Context) (app.CompositeStatus, error) {
		var obj app.InstallWorkflowStep
		if err := a.getStatus(ctx, &obj, req.ID); err != nil {
			return app.CompositeStatus{}, err
		}

		return obj.Status, nil
	}

	return a.updateStatus(ctx, &obj, req.Status, getter)
}

// @temporal-gen activity
func (a *Activities) PkgStatusUpdateInstallStackVersionStatus(ctx context.Context, req UpdateStatusRequest) error {
	obj := app.InstallStackVersion{
		ID: req.ID,
	}

	getter := func(ctx context.Context) (app.CompositeStatus, error) {
		var obj app.InstallStackVersion
		if err := a.getStatus(ctx, &obj, req.ID); err != nil {
			return app.CompositeStatus{}, err
		}

		return obj.Status, nil
	}

	return a.updateStatus(ctx, &obj, req.Status, getter)
}

func (a *Activities) updateStatus(ctx context.Context, obj any, status app.CompositeStatus, statusGetter func(ctx context.Context) (app.CompositeStatus, error)) error {
	createdBy, err := cctx.AccountIDFromContext(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to get created by")
	}

	status.CreatedByID = createdBy
	status.CreatedAtTS = time.Now().Unix()

	existingStatus, err := statusGetter(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to get existing status")
	}
	history := existingStatus.History
	existingStatus.History = nil
	status.History = append(history, existingStatus)

	res := a.db.WithContext(ctx).Model(obj).Updates(
		map[string]any{
			"status": status,
		})
	if res.Error != nil {
		return errors.Wrap(res.Error, "unable to update")
	}
	if res.RowsAffected < 1 {
		return errors.New("no object found to update")
	}
	return nil
}
