package activities

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CreateInstallWorkflowStepRequest struct {
	InstallWorkflowID string                        `json:"install_workflow_id"`
	InstallID         string                        `json:"install_id"`
	OwnerID           string                        `json:"owner_id"`
	OwnerType         string                        `json:"owner_type"`
	Status            app.CompositeStatus           `json:"status"`
	Name              string                        `json:"name"`
	Signal            app.Signal                    `json:"signal"`
	Idx               int                           `json:"idx"`
	ExecutionType     app.WorkflowStepExecutionType `json:"execution_type"`
	Metadata          pgtype.Hstore                 `json:"metadata"`
}

// @temporal-gen activity
func (a *Activities) CreateInstallWorkflowStep(ctx context.Context, req CreateInstallWorkflowStepRequest) error {
	step := &app.WorkflowStep{
		InstallWorkflowID: req.InstallWorkflowID,
		InstallID:         req.InstallID,
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
