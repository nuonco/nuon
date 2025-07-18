package activities

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CreateFlowStepRequest struct {
	FlowID        string                        `json:"flow_id" validate:"required"`
	OwnerID       string                        `json:"owner_id" validate:"required"`
	OwnerType     string                        `json:"owner_type" validate:"required"`
	Status        app.CompositeStatus           `json:"status"`
	Name          string                        `json:"name"`
	Signal        app.Signal                    `json:"signal"`
	Idx           int                           `json:"idx"`
	ExecutionType app.WorkflowStepExecutionType `json:"execution_type"`
	Metadata      pgtype.Hstore                 `json:"metadata"`
	Retryable     bool                          `json:"retryable"`
}

// @temporal-gen activity
func (a *Activities) PkgWorkflowsFlowCreateFlowStep(ctx context.Context, req CreateFlowStepRequest) (*app.WorkflowStep, error) {
	step := &app.WorkflowStep{
		InstallWorkflowID: req.FlowID,
		OwnerID:           req.OwnerID,
		OwnerType:         req.OwnerType,
		Status:            req.Status,
		Name:              req.Name,
		Signal:            req.Signal,
		Idx:               req.Idx,
		ExecutionType:     req.ExecutionType,
		Metadata:          req.Metadata,
		Retryable:         req.Retryable,
	}

	if res := a.db.WithContext(ctx).Create(step); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create step")
	}

	return step, nil
}
