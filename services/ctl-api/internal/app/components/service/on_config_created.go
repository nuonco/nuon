package service

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals"
	configcreated "github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals/v2/configcreated"
	updatecomponenttype "github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals/v2/updatecomponenttype"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

// onConfigCreated handles the component config creation signal dispatch.
// When queues are enabled, it enqueues configcreated and updatecomponenttype queue signals.
// Otherwise, it falls back to the event loop.
func (s *service) onConfigCreated(ctx *gin.Context, cmpID string, componentType app.ComponentType) error {
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
			Signal: &configcreated.Signal{
				ComponentID: cmpID,
			},
		}); err != nil {
			return fmt.Errorf("unable to enqueue config created signal: %w", err)
		}

		if _, err := s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
			QueueID:   q.ID,
			OwnerID:   cmpID,
			OwnerType: "components",
			Signal: &updatecomponenttype.Signal{
				ComponentID:   cmpID,
				ComponentType: componentType,
			},
		}); err != nil {
			return fmt.Errorf("unable to enqueue update component type signal: %w", err)
		}
	} else {
		s.evClient.Send(ctx, cmpID, &signals.Signal{
			Type: signals.OperationConfigCreated,
		})
		s.evClient.Send(ctx, cmpID, &signals.Signal{
			Type:          signals.OperationUpdateComponentType,
			ComponentType: componentType,
		})
	}

	return nil
}
