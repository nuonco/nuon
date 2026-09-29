package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type workflowStatusRow struct {
	ID     string
	Status app.CompositeStatus `gorm:"type:jsonb;serializer:json"`
}

func (workflowStatusRow) TableName() string {
	return (&app.Workflow{}).TableName()
}

type WorkflowCompletionOutcome struct {
	Status                 app.Status `json:"status"`
	StatusHumanDescription string     `json:"status_human_description,omitempty"`
}

func (o *WorkflowCompletionOutcome) HumanDescription() string {
	if o.StatusHumanDescription != "" {
		return o.StatusHumanDescription
	}
	switch o.Status {
	case app.StatusError:
		return "workflow failed"
	case app.StatusCancelled:
		return "workflow cancelled"
	}
	return ""
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
// @as-wrapper
// @by-field workflowID
// @local
func (a *Activities) workflowCompletionOutcome(ctx context.Context, workflowID string) (*WorkflowCompletionOutcome, error) {
	var flw workflowStatusRow
	if err := a.db.WithContext(ctx).
		Select("status").
		Where(workflowStatusRow{ID: workflowID}).
		First(&flw).Error; err != nil {
		return nil, fmt.Errorf("unable to load workflow status for completion outcome: %w", err)
	}

	return &WorkflowCompletionOutcome{
		Status:                 flw.Status.Status,
		StatusHumanDescription: flw.Status.StatusHumanDescription,
	}, nil
}
