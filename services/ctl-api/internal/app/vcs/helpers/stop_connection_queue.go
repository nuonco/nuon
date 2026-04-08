package helpers

import (
	"context"
	"fmt"
)

// StopConnectionQueue stops the queue associated with a VCS connection.
func (h *Helpers) StopConnectionQueue(ctx context.Context, queueID string) error {
	if err := h.queueClient.Stop(ctx, queueID); err != nil {
		return fmt.Errorf("unable to stop vcs connection queue: %w", err)
	}
	return nil
}
