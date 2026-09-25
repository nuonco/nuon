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

// WorkflowCompletionOutcome carries the workflow row's domain outcome. Resident
// flows complete their queue signal independently of the workflow row, so
// completion side effects (parent callbacks, lifecycle notifications) must be
// gated on this rather than on the queue signal's transport status.
type WorkflowCompletionOutcome struct {
	Status                 app.Status `json:"status"`
	StatusHumanDescription string     `json:"status_human_description,omitempty"`
}

// HumanDescription returns the workflow row's human description, falling back
// to a generic per-status phrase when the writer left it empty. Several status
// writers set only Status; without the fallback a parent callback or user
// notification would carry an empty error message.
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
