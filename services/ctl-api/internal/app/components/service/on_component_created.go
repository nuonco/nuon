package service

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals"
	createdsignal "github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals/v2/created"
	polldependencies "github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals/v2/polldependencies"
	provision "github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals/v2/provision"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

// onComponentCreated handles the component creation signal dispatch.
// When queues are enabled, it enqueues created, provision, and polldependencies queue signals.
// Otherwise, it falls back to the event loop.
func (s *service) onComponentCreated(ctx *gin.Context, cmpID string) error {
	useQueues, err := s.featuresClient.AllFeaturesEnabled(ctx, app.OrgFeatureAppBranches, app.OrgFeatureQueues)
	if err != nil {
		return fmt.Errorf("unable to check features: %w", err)
	}

	if useQueues {
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
	} else {
		s.evClient.Send(ctx, cmpID, &signals.Signal{
			Type: signals.OperationCreated,
		})
		s.evClient.Send(ctx, cmpID, &signals.Signal{
			Type: signals.OperationProvision,
		})
		s.evClient.Send(ctx, cmpID, &signals.Signal{
			Type: signals.OperationPollDependencies,
		})
	}

	return nil
}
