package activities

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CreateFlowStep struct {
	FlowID         string                        `json:"flow_id" validate:"required"`
	OwnerID        string                        `json:"owner_id" validate:"required"`
	OwnerType      string                        `json:"owner_type" validate:"required"`
	Status         app.CompositeStatus           `json:"status"`
	Name           string                        `json:"name"`
	Signal         app.Signal                    `json:"signal"`
	Idx            int                           `json:"idx"`
	ExecutionType  app.WorkflowStepExecutionType `json:"execution_type"`
	Metadata       pgtype.Hstore                 `json:"metadata"`
	Retryable      bool                          `json:"retryable"`
	Skippable      bool                          `json:"skippable"`
	GroupIdx       int                           `json:"group_idx"`
	GroupRetryIdx  int                           `json:"group_retry_idx"`
	StepTargetType string                        `json:"step_target_type"`
	StepTargetID   string                        `json:"step_target_id"`
}

type CreateFlowStepsRequest struct {
	Steps []CreateFlowStep `json:"steps" validate:"required"`
}

// @temporal-gen activity
func (a *Activities) PkgWorkflowsFlowCreateFlowSteps(ctx context.Context, reqs CreateFlowStepsRequest) ([]*app.WorkflowStep, error) {
	if len(reqs.Steps) == 0 {
		return []*app.WorkflowStep{}, nil
	}

	var existingSteps []*app.WorkflowStep
	if res := a.db.WithContext(ctx).
		Where(app.WorkflowStep{InstallWorkflowID: reqs.Steps[0].FlowID}).
		Order("group_idx, group_retry_idx, idx, created_at asc").
		Find(&existingSteps); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get workflow steps")
	}

	if len(existingSteps) == len(reqs.Steps) {
		// all steps already exist, nothing to do
		return existingSteps, nil
	}

	steps := make([]*app.WorkflowStep, 0, len(reqs.Steps))
	for _, req := range reqs.Steps {
		step := app.WorkflowStep{
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
			Skippable:         req.Skippable,
			GroupIdx:          req.GroupIdx,
			GroupRetryIdx:     req.GroupRetryIdx,
			StepTargetType:    req.StepTargetType,
			StepTargetID:      req.StepTargetID,
		}
		steps = append(steps, &step)
	}

	if res := a.db.WithContext(ctx).Create(steps); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create step")
	}

	return steps, nil
}
