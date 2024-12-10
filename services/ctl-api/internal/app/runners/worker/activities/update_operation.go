package activities

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type UpdateOperationRequest struct {
	OperationID string                    `validate:"required"`
	Status      app.RunnerOperationStatus `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) UpdateOperation(ctx context.Context, req UpdateOperationRequest) error {
	currentOperation := app.RunnerOperation{
		ID: req.OperationID,
	}

	res := a.db.WithContext(ctx).
		Model(&currentOperation).
		Updates(app.RunnerOperation{
			Status: req.Status,
		})
	if res.Error != nil {
		return errors.Wrap(res.Error, "unable to update runner operation")
	}
	if res.RowsAffected != 1 {
		return errors.Wrap(gorm.ErrRecordNotFound, "no operation found")
	}

	return nil
}
