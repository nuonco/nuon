package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"gorm.io/gorm/clause"
)

type CreateStepApprovalInput struct {
	OwnerID   string                       `json:"owner_id" validate:"required"`
	OwnerType string                       `json:"owner_type" validate:"required"`
	StepID    string                       `json:"step_id" validate:"required"`
	Type      app.WorkflowStepApprovalType `json:"type" validate:"required"`
	Plan      string                       `json:"plan"`
}

type CreateStepApprovalOutput struct {
	ID string `json:"id"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) CreateStepApproval(ctx context.Context, req *CreateStepApprovalInput) (*CreateStepApprovalOutput, error) {
	if err := a.db.WithContext(ctx).
		Where(app.WorkflowStepApproval{InstallWorkflowStepID: req.StepID}).
		Clauses(clause.Returning{}).
		Delete(&app.WorkflowStepApproval{}).Error; err != nil {
		return nil, fmt.Errorf("unable to soft-delete existing approval: %w", err)
	}

	sa := app.WorkflowStepApproval{
		InstallWorkflowStepID: req.StepID,
		OwnerType:             req.OwnerType,
		OwnerID:               req.OwnerID,
		Contents:              req.Plan,
		Type:                  req.Type,
	}

	if err := a.db.WithContext(ctx).Create(&sa).Error; err != nil {
		return nil, fmt.Errorf("unable to create step approval: %w", err)
	}

	return &CreateStepApprovalOutput{ID: sa.ID}, nil
}
