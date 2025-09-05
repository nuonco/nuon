package activities

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CreateInstallWorkflowStepRequest struct {
	InstallWorkflowID string                        `json:"install_workflow_id" validate:"required"`
	InstallID         string                        `json:"install_id" validate:"required"`
	OwnerID           string                        `json:"owner_id" validate:"required"`
	OwnerType         string                        `json:"owner_type" validate:"required"`
	Status            app.CompositeStatus           `json:"status" validate:"required"`
	Name              string                        `json:"name" validate:"required"`
	Signal            app.Signal                    `json:"signal" validate:"required"`
	Idx               int                           `json:"idx" validate:"required"`
	ExecutionType     app.WorkflowStepExecutionType `json:"execution_type" validate:"required"`
	Metadata          pgtype.Hstore                 `json:"metadata" validate:"required"`
}

// @temporal-gen activity
func (a *Activities) CreateInstallWorkflowStep(ctx context.Context, req CreateInstallWorkflowStepRequest) error {
	step := &app.WorkflowStep{
		InstallWorkflowID: req.InstallWorkflowID,
		OwnerID:           req.OwnerID,
		OwnerType:         req.OwnerType,
		Status:            req.Status,
		Name:              req.Name,
		Signal:            req.Signal,
		Idx:               req.Idx,
		ExecutionType:     req.ExecutionType,
		Metadata:          req.Metadata,
	}

	if res := a.db.WithContext(ctx).Create(step); res.Error != nil {
		return errors.Wrap(res.Error, "unable to create steps")
	}

	return nil
}
