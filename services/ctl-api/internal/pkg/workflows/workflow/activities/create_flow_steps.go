package activities

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

type CreateFlowStep struct {
	FlowID         string                        `json:"flow_id" validate:"required"`
	OwnerID        string                        `json:"owner_id" validate:"required"`
	OwnerType      string                        `json:"owner_type" validate:"required"`
	Status         app.CompositeStatus           `json:"status"`
	Name           string                        `json:"name"`
	Idx            int                           `json:"idx"`
	ExecutionType  app.WorkflowStepExecutionType `json:"execution_type"`
	Metadata       pgtype.Hstore                 `json:"metadata"`
	Retryable      bool                          `json:"retryable"`
	Skippable      bool                          `json:"skippable"`
	GroupIdx       int                           `json:"group_idx"`
	GroupRetryIdx  int                           `json:"group_retry_idx"`
	StepTargetType string                        `json:"step_target_type"`
	StepTargetID   string                        `json:"step_target_id"`

	// TODO(jm): this is being deprecated
	Signal      *app.Signal          `json:"signal"`
	QueueSignal *signaldb.SignalData `json:"queue_signal"`
}

type CreateFlowStepsRequest struct {
	Steps []CreateFlowStep `json:"steps" validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) PkgWorkflowsFlowCreateFlowSteps(ctx context.Context, reqs CreateFlowStepsRequest) ([]*app.WorkflowStep, error) {
	if len(reqs.Steps) == 0 {
		return []*app.WorkflowStep{}, nil
	}

	steps := make([]*app.WorkflowStep, 0, len(reqs.Steps))
	for _, req := range reqs.Steps {
		step := app.WorkflowStep{
			InstallWorkflowID: req.FlowID,
			OwnerID:           req.OwnerID,
			OwnerType:         req.OwnerType,
			Status:            req.Status,
			Name:              req.Name,
			Idx:               req.Idx,
			ExecutionType:     req.ExecutionType,
			Metadata:          req.Metadata,
			Retryable:         req.Retryable,
			Skippable:         req.Skippable,
			GroupIdx:          req.GroupIdx,
			GroupRetryIdx:     req.GroupRetryIdx,
			StepTargetType:    req.StepTargetType,
			StepTargetID:      req.StepTargetID,
			Signal:            req.Signal,
			QueueSignal:       req.QueueSignal,
		}
		steps = append(steps, &step)
	}

	if res := a.db.WithContext(ctx).Create(steps); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create step")
	}

	return steps, nil
}
