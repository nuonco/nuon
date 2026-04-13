package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type IncrementQueueSignalExecutionCountRequest struct {
	QueueSignalID string `json:"queue_signal_id" validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) IncrementQueueSignalExecutionCount(ctx context.Context, req *IncrementQueueSignalExecutionCountRequest) error {
	res := a.db.WithContext(ctx).
		Model(&app.QueueSignal{}).
		Where("id = ?", req.QueueSignalID).
		Update("execution_count", gorm.Expr("execution_count + 1"))
	if res.Error != nil {
		return generics.TemporalGormError(res.Error, fmt.Sprintf("unable to increment execution count for queue signal %s", req.QueueSignalID))
	}

	return nil
}
