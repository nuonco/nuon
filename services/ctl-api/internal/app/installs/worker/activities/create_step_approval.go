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
	StepID      string `validate:"required"`
	Plan        string
	Type        app.InstallWorkflowStepApprovalType `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) CreateStepApproval(ctx context.Context, req *CreateStepApprovalRequest) (*app.InstallWorkflowStepApproval, error) {
	sa := app.InstallWorkflowStepApproval{
		InstallWorkflowStepID: req.StepID,
		OwnerType:             req.OwnerType,
		OwnerID:               req.OwnerID,
		Contents:              req.Plan,
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
