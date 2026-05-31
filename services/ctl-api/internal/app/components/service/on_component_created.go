package service

import (
	"fmt"

	"github.com/gin-gonic/gin"

	createdsignal "github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals/created"
	polldependencies "github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals/polldependencies"
	provision "github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals/provision"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

// onComponentCreated handles the component creation signal dispatch.
func (s *service) onComponentCreated(ctx *gin.Context, cmpID string) error {
	q, err := s.queueClient.GetQueueByOwner(ctx, cmpID, "components")
	if err != nil {
		return fmt.Errorf("unable to get component queue: %w", err)
	}

	if _, err := s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID:   q.ID,
		OwnerID:   cmpID,
		OwnerType: "components",
		Signal: &createdsignal.Signal{
			ComponentID: cmpID,
		},
	}); err != nil {
		return fmt.Errorf("unable to enqueue created signal: %w", err)
	}

	if _, err := s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID:   q.ID,
		OwnerID:   cmpID,
		OwnerType: "components",
		Signal: &provision.Signal{
			ComponentID: cmpID,
		},
	}); err != nil {
		return fmt.Errorf("unable to enqueue provision signal: %w", err)
	}

	if _, err := s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID:   q.ID,
		OwnerID:   cmpID,
		OwnerType: "components",
		Signal: &polldependencies.Signal{
			ComponentID: cmpID,
		},
	}); err != nil {
		return fmt.Errorf("unable to enqueue poll dependencies signal: %w", err)
	}

	return nil
}
