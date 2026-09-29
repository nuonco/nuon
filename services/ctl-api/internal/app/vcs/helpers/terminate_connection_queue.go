package helpers

import (
	"context"
	"fmt"
)

func (h *Helpers) TerminateConnectionQueue(ctx context.Context, queueID string) error {
	if err := h.queueClient.Terminate(ctx, queueID); err != nil {
		return fmt.Errorf("unable to terminate vcs connection queue: %w", err)
	}
	return nil
}
