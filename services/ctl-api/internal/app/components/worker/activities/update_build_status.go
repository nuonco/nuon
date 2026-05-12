package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

type UpdateBuildStatus struct {
	BuildID           string                   `validate:"required"`
	Status            app.ComponentBuildStatus `validate:"required"`
	StatusDescription string                   `validate:"required"`

	// CompositeError is an optional typed, structured error to attach to the
	// build alongside the status update. When non-nil it is persisted to the
	// build's composite_error JSONB column.
	CompositeError *compositeerrors.CompositeErrorData
}

// @temporal-gen-v2 activity
func (a *Activities) UpdateBuildStatus(ctx context.Context, req UpdateBuildStatus) error {
	currentApp := app.ComponentBuild{
		ID: req.BuildID,
	}
	updates := app.ComponentBuild{
		Status:            req.Status,
		StatusDescription: req.StatusDescription,
	}
	if req.CompositeError != nil {
		updates.CompositeError = req.CompositeError
	}
	res := a.db.WithContext(ctx).Model(&currentApp).Updates(updates)
	if res.Error != nil {
		return fmt.Errorf("unable to update build: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return fmt.Errorf("no build found: %s %w", req.BuildID, gorm.ErrRecordNotFound)
	}

	return nil
}
