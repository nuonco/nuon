package service

import (
	"context"
	"fmt"

	createdsignal "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/created"
	polldependencies "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/polldependencies"
	provision "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/provision"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

// onAppCreated enqueues the created, provision and poll-dependencies signals on
// the app queue so a newly created app gets provisioned.
func (s *service) onAppCreated(ctx context.Context, appID string) error {
	q, err := s.queueClient.GetQueueByOwner(ctx, appID, "apps")
	if err != nil {
		return fmt.Errorf("unable to get app queue: %w", err)
	}

	if _, err := s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID:   q.ID,
		OwnerID:   appID,
		OwnerType: "apps",
		Signal:    &createdsignal.Signal{AppID: appID},
	}); err != nil {
		return fmt.Errorf("unable to enqueue created signal: %w", err)
	}

	if _, err := s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID:   q.ID,
		OwnerID:   appID,
		OwnerType: "apps",
		Signal:    &provision.Signal{AppID: appID},
	}); err != nil {
		return fmt.Errorf("unable to enqueue provision signal: %w", err)
	}

	if _, err := s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID:   q.ID,
		OwnerID:   appID,
		OwnerType: "apps",
		Signal:    &polldependencies.Signal{AppID: appID},
	}); err != nil {
		return fmt.Errorf("unable to enqueue poll dependencies signal: %w", err)
	}

	return nil
}
