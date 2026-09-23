package activities

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/plancounts"
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

	counts, state, err := plancounts.Counts(req.Type, req.Plan)
	if err != nil {
		a.l.Warn("unable to summarize approval plan",
			zap.String("step_id", req.StepID),
			zap.String("approval_type", string(req.Type)),
			zap.Error(err))
	}
	sa.SetChanges(counts, state)

	if err := a.db.WithContext(ctx).Create(&sa).Error; err != nil {
		return nil, fmt.Errorf("unable to create step approval: %w", err)
	}

	return &CreateStepApprovalOutput{ID: sa.ID}, nil
}
