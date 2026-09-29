package helpers

import (
	"context"
	"fmt"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections/signals/reverify"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuenames"
)

func (h *Helpers) EnsureConnectionQueue(ctx context.Context, connection *app.CloudConnection) (*app.Queue, error) {
	q, err := h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OwnerID: connection.ID, OwnerType: queuenames.OwnerCloudConnections, Namespace: "orgs",
		Name: "cloud-connection-" + connection.ID, MaxInFlight: 1, MaxDepth: 5,
	})
	if err != nil {
		return nil, fmt.Errorf("ensure cloud connection queue: %w", err)
	}
	if err := h.db.WithContext(ctx).Model(q).Select("idle_timeout").Updates(app.Queue{IdleTimeout: int64(5 * time.Second)}).Error; err != nil {
		return nil, err
	}
	return q, nil
}

func (h *Helpers) EnqueueVerification(ctx context.Context, connection *app.CloudConnection) error {
	q, err := h.EnsureConnectionQueue(ctx, connection)
	if err != nil {
		return err
	}
	_, err = h.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID: q.ID, OwnerID: connection.ID, OwnerType: queuenames.OwnerCloudConnections,
		Signal: &reverify.Signal{CloudConnectionID: connection.ID, OnDemand: true},
	})
	return err
}

func (h *Helpers) TerminateConnectionQueue(ctx context.Context, connectionID string) error {
	var queues []app.Queue
	if err := h.db.WithContext(ctx).Where(app.Queue{OwnerID: connectionID, OwnerType: queuenames.OwnerCloudConnections}).Find(&queues).Error; err != nil {
		return err
	}
	for _, q := range queues {
		if err := h.queueClient.TerminateStrict(ctx, q.ID); err != nil {
			return err
		}
	}
	return nil
}
