package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

const GeneralSignalsQueueName = "general-signals"

// EnsureGeneralQueue creates the general-signals queue if it doesn't already exist.
// Safe to call multiple times — queueClient.Create is idempotent.
func (h *Helpers) EnsureGeneralQueue(ctx context.Context) (*app.Queue, error) {
	q, err := h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OwnerID:     "general",
		OwnerType:   "general",
		Namespace:   "general",
		Name:        GeneralSignalsQueueName,
		MaxInFlight: 5,
		MaxDepth:    50,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to ensure general-signals queue: %w", err)
	}
	return q, nil
}
