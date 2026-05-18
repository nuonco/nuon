package activities

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type ClearOrgQueuesRequest struct {
	OrgID string `validate:"required"`
}

type ClearOrgQueuesResponse struct {
	QueuesCleared    int `json:"queues_cleared"`
	SignalsCancelled int `json:"signals_cancelled"`
}

// @temporal-gen-v2 activity
// @by-field OrgID
func (a *Activities) ClearOrgQueues(ctx context.Context, req ClearOrgQueuesRequest) (*ClearOrgQueuesResponse, error) {
	var queues []app.Queue
	if res := a.db.WithContext(ctx).
		Where("org_id = ?", req.OrgID).
		Find(&queues); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to list org queues")
	}

	cleared := 0
	signalsCancelled := 0

	for _, q := range queues {
		var signals []app.QueueSignal
		if res := a.db.WithContext(ctx).
			Where(app.QueueSignal{QueueID: q.ID}).
			Find(&signals); res.Error != nil {
			continue
		}

		for _, qs := range signals {
			if isTerminalStatus(qs.Status.Status) {
				continue
			}

			cancelledStatus := app.CompositeStatus{
				CreatedAtTS:            time.Now().Unix(),
				Status:                 app.StatusCancelled,
				StatusHumanDescription: "cancelled by clear-org-queues",
				Metadata:               map[string]any{"cancelled_by": "clear-org-queues"},
			}
			if res := a.db.WithContext(ctx).
				Model(&app.QueueSignal{}).
				Where("id = ?", qs.ID).
				Update("status", cancelledStatus); res.Error != nil {
				continue
			}
			signalsCancelled++
		}
		cleared++
	}

	return &ClearOrgQueuesResponse{
		QueuesCleared:    cleared,
		SignalsCancelled: signalsCancelled,
	}, nil
}

func isTerminalStatus(s app.Status) bool {
	switch s {
	case app.StatusSuccess, app.StatusCancelled, app.StatusDiscarded, app.StatusError:
		return true
	}
	return false
}
