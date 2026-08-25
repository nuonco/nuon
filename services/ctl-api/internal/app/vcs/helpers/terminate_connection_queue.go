package helpers

import (
	"context"
	"fmt"
)

// TerminateConnectionQueue terminates queues so queue and emitter rows are deleted too
func (h *Helpers) TerminateConnectionQueue(ctx context.Context, queueID string) error {
	if err := h.queueClient.Terminate(ctx, queueID); err != nil {
		return fmt.Errorf("unable to terminate vcs connection queue: %w", err)
	}
	return nil
}
