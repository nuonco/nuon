package helpers

import (
	"context"
	"fmt"
)

func (h *Helpers) StopProcessQueue(ctx context.Context, queueID string) error {
	emitters, err := h.emitterClient.GetEmittersByQueueID(ctx, queueID)
	if err != nil {
		return fmt.Errorf("unable to get emitters for queue %s: %w", queueID, err)
	}

	for _, em := range emitters {
		if _, err := h.emitterClient.StopEmitter(ctx, em.ID); err != nil {
			return fmt.Errorf("unable to stop emitter %s: %w", em.ID, err)
		}
	}

	return nil
}
