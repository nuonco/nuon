package service

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	configcreated "github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals/configcreated"
	updatecomponenttype "github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals/updatecomponenttype"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

// onConfigCreated handles the component config creation signal dispatch.
func (s *service) onConfigCreated(ctx *gin.Context, cmpID string, componentType app.ComponentType) error {
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

	return nil
}
