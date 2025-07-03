package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CreateStepApprovalRequest struct {
	OwnerID   string `validate:"required"`
	OwnerType string `validate:"required"`

	RunnerJobID string
	StepID      string                       `validate:"required"`
	Type        app.WorkflowStepApprovalType `validate:"required"`

	Plan string
}

// @temporal-gen activity
func (a *Activities) CreateStepApproval(ctx context.Context, req *CreateStepApprovalRequest) (*app.WorkflowStepApproval, error) {
	plan := req.Plan
	if req.Plan == "" {
		job, err := a.GetJob(ctx, &GetJobRequest{
			ID: req.RunnerJobID,
		})
		if err != nil {
			return nil, errors.Wrap(err, "unable to get job")
		}

		plan = string(job.Execution.Result.ContentsDisplay)
	}

	sa := app.WorkflowStepApproval{
		InstallWorkflowStepID: req.StepID,
		OwnerType:             req.OwnerType,
		OwnerID:               req.OwnerID,
		Contents:              plan,
		Type:                  req.Type,
	}

	// workflows polymorphic step approvals do not have a runner job ID
	if req.RunnerJobID != "" {
		sa.RunnerJobID = generics.ToPtr(req.RunnerJobID)
	}

	res := a.db.WithContext(ctx).Create(&sa)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create step approval")
	}

	return &sa, nil
}
