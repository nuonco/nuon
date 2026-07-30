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

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
// @as-wrapper
// @by-field workflowID
// @local
func (a *Activities) holdCompletionCallbacks(ctx context.Context, workflowID string) (bool, error) {
	var flw workflowStatusRow
	if err := a.db.WithContext(ctx).
		Select("status").
		Where(workflowStatusRow{ID: workflowID}).
		First(&flw).Error; err != nil {
		return false, fmt.Errorf("unable to load workflow status for completion callbacks: %w", err)
	}

	return flw.Status.Status == app.StatusFailedPendingRetry, nil
}
