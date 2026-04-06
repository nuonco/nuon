package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

// CreateSubscriptionQueue creates a queue for the given VCS webhook subscription.
// This queue processes github_event signals that fan out to VCS connections.
func (h *Helpers) CreateSubscriptionQueue(ctx context.Context, sub *app.VCSWebhookSubscription) (*app.Queue, error) {
	q, err := h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OwnerID:     sub.ID,
		OwnerType:   "vcs_webhook_subscriptions",
		Namespace:   vcsTemporalNamespace,
		Name:        fmt.Sprintf("vcs-webhook-subscription-%s", sub.ID),
		MaxInFlight: 2,
		MaxDepth:    50,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create webhook subscription queue: %w", err)
	}

	return q, nil
}
