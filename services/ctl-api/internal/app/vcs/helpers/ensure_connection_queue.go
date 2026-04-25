package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/signals/v2/healthcheck"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	emitterclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/client"
)

// EnsureConnectionQueue creates the VCS connection queue if it doesn't exist.
// Safe to call multiple times — queue creation is idempotent, and the emitter
// is only created if it doesn't already exist.
func (h *Helpers) EnsureConnectionQueue(ctx context.Context, vcsConn *app.VCSConnection) error {
	queueName := fmt.Sprintf("vcs-connection-%s", vcsConn.ID)

	q, err := h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OwnerID:     vcsConn.ID,
		OwnerType:   "vcs_connections",
		Namespace:   vcsTemporalNamespace,
		Name:        queueName,
		MaxInFlight: 1,
		MaxDepth:    5,
	})
	if err != nil {
		return fmt.Errorf("unable to ensure vcs connection queue: %w", err)
	}

	// Only create the emitter if it doesn't exist.
	emitterName := fmt.Sprintf("vcs-connection-%s-health-check", vcsConn.ID)
	var existing app.QueueEmitter
	if res := h.db.WithContext(ctx).
		Where(app.QueueEmitter{QueueID: q.ID, Name: emitterName}).
		First(&existing); res.Error == nil {
		return nil
	}

	if _, err := h.emitterClient.CreateEmitter(ctx, &emitterclient.CreateEmitterRequest{
		QueueID:      q.ID,
		Name:         emitterName,
		Description:  "Periodic VCS connection health check",
		Mode:         app.QueueEmitterModeCron,
		CronSchedule: "*/5 * * * *",
		SignalType:   healthcheck.SignalType,
		SignalTemplate: &healthcheck.Signal{
			VCSConnectionID: vcsConn.ID,
		},
	}); err != nil {
		return fmt.Errorf("unable to create vcs health check emitter: %w", err)
	}

	return nil
}
