package adminops

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/general/signals/promotion"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

const enqueueSourcePromotion = "admin-promotion"

type PromoteResult struct {
	Tag           string `json:"tag"`
	QueueSignalID string `json:"queue_signal_id"`
}

func (o *Ops) Promote(ctx context.Context, tag string) (*PromoteResult, error) {
	q, err := o.generalHelpers.EnsureGeneralQueue(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to ensure general queue: %w", err)
	}

	var resp *queue.EnqueueResponse
	if err := o.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		resp, err = o.queueClient.EnqueueSignalInTransaction(ctx, tx, &queueclient.EnqueueSignalRequest{
			QueueID: q.ID,
			Signal:  &promotion.Signal{Tag: tag},
		})
		return err
	}); err != nil {
		return nil, fmt.Errorf("unable to enqueue promotion signal: %w", err)
	}

	if err := o.enqueuer.EnqueueInline(ctx, resp.ID, enqueueSourcePromotion); err != nil {
		return nil, fmt.Errorf("unable to dispatch promotion signal: %w", err)
	}

	return &PromoteResult{
		Tag:           tag,
		QueueSignalID: resp.ID,
	}, nil
}
