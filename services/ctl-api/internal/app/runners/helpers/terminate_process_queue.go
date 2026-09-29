package helpers

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (h *Helpers) TerminateProcessQueue(ctx context.Context, runnerID, processID string) error {
	queueName := fmt.Sprintf("runner-process-%s", processID)

	var q app.Queue
	if res := h.db.WithContext(ctx).
		Where(&app.Queue{
			OwnerID:   runnerID,
			OwnerType: "runners",
			Name:      queueName,
		}).First(&q); res.Error != nil {
		return fmt.Errorf("unable to find process queue %s: %w", queueName, res.Error)
	}

	if err := h.queueClient.Terminate(ctx, q.ID); err != nil {
		return fmt.Errorf("unable to terminate process queue: %w", err)
	}

	return nil
}

func (h *Helpers) TerminateProcessQueueStrict(ctx context.Context, runnerID, processID string) error {
	queueName := fmt.Sprintf("runner-process-%s", processID)

	var q app.Queue
	if res := h.db.WithContext(ctx).
		Where(&app.Queue{
			OwnerID:   runnerID,
			OwnerType: "runners",
			Name:      queueName,
		}).First(&q); res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil
		}
		return fmt.Errorf("unable to find process queue %s: %w", queueName, res.Error)
	}

	if err := h.queueClient.TerminateStrict(ctx, q.ID); err != nil {
		return fmt.Errorf("unable to terminate process queue: %w", err)
	}

	return nil
}
