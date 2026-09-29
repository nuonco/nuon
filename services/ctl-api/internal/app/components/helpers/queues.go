package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuenames"
)

const (
	ComponentWorkflowStepsQueueName = queuenames.ComponentWorkflowStepsQueueName
)

type ComponentQueueIDs struct {
	DefaultQueueID       string `json:"default_queue_id"`
	WorkflowStepsQueueID string `json:"workflow_steps_queue_id"`
}

func (h *Helpers) EnsureComponentQueues(ctx context.Context, componentID string) (*ComponentQueueIDs, error) {
	ownerType := plugins.TableName(h.db, app.Component{})
	specs, ok := queuenames.Specs(queuenames.OwnerComponents)
	if !ok {
		return nil, fmt.Errorf("component queue specs are not registered")
	}

	ids := &ComponentQueueIDs{}
	for _, spec := range specs {
		existing, err := h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
			OwnerID:     componentID,
			OwnerType:   ownerType,
			Namespace:   "components",
			Name:        spec.Name,
			MaxInFlight: spec.MaxInFlight,
			MaxDepth:    spec.MaxDepth,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to ensure %s queue for component %s: %w", spec.Name, componentID, err)
		}
		if spec.Name == queuenames.ComponentDefaultQueueName {
			ids.DefaultQueueID = existing.ID
		} else if spec.Name == queuenames.ComponentWorkflowStepsQueueName {
			ids.WorkflowStepsQueueID = existing.ID
		}
	}

	return ids, nil
}

func (h *Helpers) GetComponentQueueIDs(ctx context.Context, componentID string) (*ComponentQueueIDs, error) {
	ownerType := plugins.TableName(h.db, app.Component{})

	defaultQueue, err := h.queueClient.GetDefaultQueueByOwner(ctx, componentID, ownerType)
	if err != nil {
		return nil, fmt.Errorf("unable to get default queue for component %s: %w", componentID, err)
	}

	stepsQueue, err := h.queueClient.GetQueueByOwnerAndName(ctx, componentID, ownerType, ComponentWorkflowStepsQueueName)
	if err != nil {
		return nil, fmt.Errorf("unable to get %s queue for component %s: %w", ComponentWorkflowStepsQueueName, componentID, err)
	}

	return &ComponentQueueIDs{
		DefaultQueueID:       defaultQueue.ID,
		WorkflowStepsQueueID: stepsQueue.ID,
	}, nil
}
