package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuenames"
)

const GeneralSignalsQueueName = queuenames.GeneralSignalsQueueName

// EnsureGeneralQueue creates the general-signals queue if it doesn't already exist.
// Safe to call multiple times — queueClient.Create is idempotent.
func (h *Helpers) EnsureGeneralQueue(ctx context.Context) (*app.Queue, error) {
	spec, ok := queuenames.SpecByName(queuenames.OwnerGeneral, queuenames.GeneralSignalsQueueName)
	if !ok {
		return nil, fmt.Errorf("general signals queue is not registered")
	}
	q, err := h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OwnerID:     "general",
		OwnerType:   "general",
		Namespace:   "general",
		Name:        spec.Name,
		MaxInFlight: spec.MaxInFlight,
		MaxDepth:    spec.MaxDepth,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to ensure general-signals queue: %w", err)
	}
	return q, nil
}
