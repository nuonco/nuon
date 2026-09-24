package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuenames"
)

// EnsureOrgQueue creates the org-signals queue if it doesn't already exist.
// Safe to call multiple times — queueClient.Create is idempotent.
func (h *Helpers) EnsureOrgQueue(ctx context.Context, orgID string) error {
	_, err := h.ensureOrgQueue(ctx, orgID)
	return err
}

func (h *Helpers) ensureOrgQueue(ctx context.Context, orgID string) (*app.Queue, error) {
	spec, ok := queuenames.SpecByName(queuenames.OwnerOrgs, queuenames.OrgSignalsQueueName)
	if !ok {
		return nil, fmt.Errorf("org signals queue is not registered")
	}
	q, err := h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OrgID:       &orgID,
		OwnerID:     orgID,
		OwnerType:   plugins.TableName(h.db, app.Org{}),
		Namespace:   "orgs",
		Name:        spec.Name,
		MaxInFlight: spec.MaxInFlight,
		MaxDepth:    spec.MaxDepth,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to ensure org-signals queue: %w", err)
	}
	return q, nil
}
