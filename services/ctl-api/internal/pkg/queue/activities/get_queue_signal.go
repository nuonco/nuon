package activities

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
)

type GetQueueSignalRequest struct {
	QueueSignalID string `validate:"required"`
}

// @temporal-gen activity
// @by-id QueueSignalID
func (a *Activities) GetQueueSignal(ctx context.Context, req *GetQueueSignalRequest) (*app.QueueSignal, error) {
	queueSignal, err := a.getQueueSignal(ctx, req.QueueSignalID)
	if err != nil {
		return nil, generics.TemporalGormError(err, "unable to get queue signal")
	}

	return queueSignal, nil
}
