package adminops

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/general/signals/promotion"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

const enqueueSourcePromotion = "admin-promotion"

type PromoteResult struct {
	Tag           string `json:"tag"`
	QueueSignalID string `json:"queue_signal_id"`
	Deduplicated  bool   `json:"deduplicated"`
}

func (o *Ops) Promote(ctx context.Context, tag string) (*PromoteResult, error) {
	q, err := o.generalHelpers.EnsureGeneralQueue(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to ensure general queue: %w", err)
	}

	return o.enqueuePromotion(ctx, q.ID, tag, "")
}

func (o *Ops) PromoteOnce(ctx context.Context, tag string) (*PromoteResult, error) {
	q, err := o.generalHelpers.EnsureGeneralQueue(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to ensure general queue: %w", err)
	}

	latest, err := o.latestPromotion(ctx, q.ID)
	if err != nil {
		return nil, err
	}

	if latest != nil && isReusablePromotion(latest, tag) {
		if err := o.dispatchPromotion(ctx, latest.ID); err != nil {
			return nil, err
		}

		return &PromoteResult{
			Tag:           tag,
			QueueSignalID: latest.ID,
			Deduplicated:  true,
		}, nil
	}

	return o.enqueuePromotion(ctx, q.ID, tag, promotionIdempotencyKey(tag, latest))
}

func (o *Ops) enqueuePromotion(ctx context.Context, queueID, tag, idempotencyKey string) (*PromoteResult, error) {
	var resp *queue.EnqueueResponse
	if err := o.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		resp, err = o.queueClient.EnqueueSignalInTransaction(ctx, tx, &queueclient.EnqueueSignalRequest{
			QueueID:        queueID,
			Signal:         &promotion.Signal{Tag: tag},
			IdempotencyKey: idempotencyKey,
		})
		return err
	}); err != nil {
		return nil, fmt.Errorf("unable to enqueue promotion signal: %w", err)
	}

	if err := o.dispatchPromotion(ctx, resp.ID); err != nil {
		return nil, err
	}

	return &PromoteResult{
		Tag:           tag,
		QueueSignalID: resp.ID,
		Deduplicated:  resp.Deduplicated,
	}, nil
}

func (o *Ops) dispatchPromotion(ctx context.Context, queueSignalID string) error {
	if err := o.enqueuer.EnqueueInline(ctx, queueSignalID, enqueueSourcePromotion); err != nil {
		return fmt.Errorf("unable to dispatch promotion signal: %w", err)
	}

	return nil
}

func (o *Ops) latestPromotion(ctx context.Context, queueID string) (*app.QueueSignal, error) {
	var latest app.QueueSignal
	res := o.db.WithContext(ctx).
		Where(app.QueueSignal{QueueID: queueID, Type: promotion.SignalType}).
		Order("created_at DESC").
		First(&latest)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get latest promotion signal: %w", res.Error)
	}

	return &latest, nil
}

func isReusablePromotion(latest *app.QueueSignal, tag string) bool {
	sig, ok := latest.Signal.Signal.(*promotion.Signal)
	if !ok || sig.Tag != tag {
		return false
	}

	switch latest.Status.Status {
	case app.StatusQueued, app.StatusPending, app.StatusInProgress, app.StatusSuccess:
		return true
	default:
		return false
	}
}

func promotionIdempotencyKey(tag string, latest *app.QueueSignal) string {
	if latest == nil {
		return fmt.Sprintf("promotion-%s", tag)
	}

	return fmt.Sprintf("promotion-%s-after-%s", tag, latest.ID)
}
