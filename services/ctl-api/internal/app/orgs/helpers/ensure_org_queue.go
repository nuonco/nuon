package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

// EnsureOrgQueue creates the org-signals queue if it doesn't already exist.
// If it exists, it restarts the queue workflow to ensure it's running.
func (h *Helpers) EnsureOrgQueue(ctx context.Context, orgID string) error {
	var existing app.Queue
	if res := h.db.WithContext(ctx).
		Where(app.Queue{OwnerID: orgID, Name: OrgSignalsQueueName}).
		First(&existing); res.Error == nil {
		return h.queueClient.Restart(ctx, existing.ID)
	}

	_, err := h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OrgID:       &orgID,
		OwnerID:     orgID,
		OwnerType:   plugins.TableName(h.db, app.Org{}),
		Namespace:   "orgs",
		Name:        OrgSignalsQueueName,
		MaxInFlight: 10,
		MaxDepth:    50,
	})
	if err != nil {
		return fmt.Errorf("unable to create org-signals queue: %w", err)
	}

	return nil
}
