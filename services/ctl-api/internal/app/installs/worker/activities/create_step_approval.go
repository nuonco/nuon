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

	RunnerJobID string                              `validate:"required"`
	StepID      string                              `validate:"required"`
	Plan        string                              `validate:"required"`
	Type        app.InstallWorkflowStepApprovalType `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) CreateStepApproval(ctx context.Context, req *CreateStepApprovalRequest) (*app.InstallWorkflowStepApproval, error) {
	sa := app.InstallWorkflowStepApproval{
		InstallWorkflowStepID: req.StepID,
		RunnerJobID:           generics.ToPtr(req.RunnerJobID),
		OwnerType:             req.OwnerType,
		OwnerID:               req.OwnerID,
		Contents:              req.Plan,
		Type:                  req.Type,
	}

	res := a.db.WithContext(ctx).Create(&sa)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create log stream")
	}

	return &sa, nil
}
