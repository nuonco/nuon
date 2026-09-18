package activities

import (
	"context"
	"fmt"

	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type EnsureSubscriptionQueueRequest struct {
	SubscriptionID string `json:"subscription_id" validate:"required"`
}

type EnsureSubscriptionQueueResponse struct {
	QueueID string `json:"queue_id"`
}

// @temporal-gen-v2 activity
func (a *Activities) EnsureSubscriptionQueue(ctx context.Context, req EnsureSubscriptionQueueRequest) (*EnsureSubscriptionQueueResponse, error) {
	q, err := a.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OwnerID:     req.SubscriptionID,
		OwnerType:   "vcs_webhook_subscriptions",
		Namespace:   "vcs",
		Name:        fmt.Sprintf("vcs-webhook-subscription-%s", req.SubscriptionID),
		MaxInFlight: 2,
		MaxDepth:    50,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create webhook subscription queue: %w", err)
	}

	return &EnsureSubscriptionQueueResponse{QueueID: q.ID}, nil
}
