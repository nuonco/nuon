package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/signals/v2/healthcheck"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	emitterclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/client"
)

// vcsTemporalNamespace mirrors the constant in vcs/worker to avoid import cycle.
const vcsTemporalNamespace = "vcs"

// CreateConnectionQueue creates a queue for the given VCS connection with a cron health check
// emitter that fires every 5 minutes, and enqueues an immediate health check signal.
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
