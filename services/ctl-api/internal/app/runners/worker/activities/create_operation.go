package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CreateOperationRequest struct {
	RunnerID      string                  `validate:"required"`
	OperationType app.RunnerOperationType `validate:"required"`
}

// @temporal-gen activity
// @schedule-to-close-timeout 5s
func (a *Activities) CreateOperationRequest(ctx context.Context, req CreateOperationRequest) (*app.RunnerOperation, error) {
	op := app.RunnerOperation{
		OpType:   req.OperationType,
		RunnerID: req.RunnerID,
		Status:   app.RunnerOperationStatusPending,
	}
	if res := a.db.WithContext(ctx).Create(&op); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create operation")
	}

	return &op, nil
}
