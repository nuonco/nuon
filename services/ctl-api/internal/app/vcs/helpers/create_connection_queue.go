package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/signals/v2/healthcheck"
	webhooksubscription "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/signals/v2/webhook_subscription"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	emitterclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/client"
)

// vcsTemporalNamespace mirrors the constant in vcs/worker to avoid import cycle.
const vcsTemporalNamespace = "vcs"

// CreateConnectionQueue creates a queue for the given VCS connection with a cron health check
// emitter that fires every 5 minutes and a fire-once webhook subscription emitter.
func (h *Helpers) CreateConnectionQueue(ctx context.Context, vcsConn *app.VCSConnection) (*app.Queue, error) {
	q, err := h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OwnerID:     vcsConn.ID,
		OwnerType:   "vcs_connections",
		Namespace:   vcsTemporalNamespace,
		Name:        fmt.Sprintf("vcs-connection-%s", vcsConn.ID),
		MaxInFlight: 1,
		MaxDepth:    5,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create vcs connection queue: %w", err)
	}

	// Cron emitter: health check every 5 minutes
	if _, err := h.emitterClient.CreateEmitter(ctx, &emitterclient.CreateEmitterRequest{
		QueueID:      q.ID,
		Name:         fmt.Sprintf("vcs-connection-%s-health-check", vcsConn.ID),
		Description:  "Periodic VCS connection health check",
		Mode:         app.QueueEmitterModeCron,
		CronSchedule: "*/5 * * * *",
		SignalType:   healthcheck.SignalType,
		SignalTemplate: &healthcheck.Signal{
			VCSConnectionID: vcsConn.ID,
		},
	}); err != nil {
		return nil, fmt.Errorf("unable to create vcs health check emitter: %w", err)
	}

	// Fire-once emitter: create webhook subscription on first queue run
	if _, err := h.emitterClient.CreateEmitter(ctx, &emitterclient.CreateEmitterRequest{
		QueueID:     q.ID,
		Name:        fmt.Sprintf("vcs-connection-%s-webhook-subscription", vcsConn.ID),
		Description: "Create webhook subscription for VCS connection",
		Mode:        app.QueueEmitterModeFireOnce,
		SignalType:  webhooksubscription.SignalType,
		SignalTemplate: &webhooksubscription.Signal{
			VCSConnectionID: vcsConn.ID,
		},
	}); err != nil {
		return nil, fmt.Errorf("unable to create vcs webhook subscription emitter: %w", err)
	}

	// Update VCS connection with queue ID
	if err := h.db.WithContext(ctx).
		Model(&app.VCSConnection{}).
		Where("id = ?", vcsConn.ID).
		Update("queue_id", q.ID).Error; err != nil {
		return nil, fmt.Errorf("unable to update vcs connection queue_id: %w", err)
	}

	return q, nil
}
