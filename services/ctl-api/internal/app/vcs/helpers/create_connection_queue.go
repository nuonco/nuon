package helpers

import (
	"context"
	"fmt"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/signals/healthcheck"
	webhooksubscription "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/signals/webhook_subscription"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	emitterclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/client"
)

// vcsTemporalNamespace mirrors the constant in vcs/worker to avoid import cycle.
const vcsTemporalNamespace = "vcs"

// CreateConnectionQueue creates a queue for the given VCS connection with a cron health check
// emitter that fires every minute, a fire-once webhook subscription emitter, and enqueues an immediate health check signal.
func (h *Helpers) CreateConnectionQueue(ctx context.Context, vcsConn *app.VCSConnection) (*app.Queue, error) {
	q, err := h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OwnerID:     vcsConn.ID,
		OwnerType:   "vcs_connections",
		Namespace:   vcsTemporalNamespace,
		Name:        fmt.Sprintf("vcs-connection-%s", vcsConn.ID),
		MaxInFlight: 1,
		MaxDepth:    5,
	})
	if err == nil {
		h.db.WithContext(ctx).Model(q).Update("idle_timeout", int64(5*time.Second))
	}
	if err != nil {
		return nil, fmt.Errorf("unable to create vcs connection queue: %w", err)
	}

	// Cron emitter: health check every minute
	if _, err := h.emitterClient.CreateEmitter(ctx, &emitterclient.CreateEmitterRequest{
		QueueID:         q.ID,
		Name:            fmt.Sprintf("vcs-connection-%s-health-check", vcsConn.ID),
		Description:     "Periodic VCS connection health check",
		Mode:            app.QueueEmitterModeCron,
		CronSchedule:    "0 * * * *",
		SignalExpiresIn: 5 * time.Minute,
		SignalType:      healthcheck.SignalType,
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

	// Enqueue an immediate health check signal
	if _, err := h.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID: q.ID,
		Signal: &healthcheck.Signal{
			VCSConnectionID: vcsConn.ID,
		},
	}); err != nil {
		return nil, fmt.Errorf("unable to enqueue initial health check: %w", err)
	}

	return q, nil
}
