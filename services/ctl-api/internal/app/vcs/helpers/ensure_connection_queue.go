package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// EnsureConnectionQueue creates the VCS connection queue if it doesn't exist.
// If it exists, restarts the queue workflow.
func (h *Helpers) EnsureConnectionQueue(ctx context.Context, vcsConn *app.VCSConnection) error {
	queueName := fmt.Sprintf("vcs-connection-%s", vcsConn.ID)

	var existing app.Queue
	if res := h.db.WithContext(ctx).
		Where(app.Queue{OwnerID: vcsConn.ID, Name: queueName}).
		First(&existing); res.Error == nil {
		return h.queueClient.Restart(ctx, existing.ID)
	}

	_, err := h.CreateConnectionQueue(ctx, vcsConn)
	return err
}
