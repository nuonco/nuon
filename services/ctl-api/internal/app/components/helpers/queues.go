package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

const (
	// ComponentWorkflowStepsQueueName is the named queue on each component
	// where execute-workflow-step signals run for component builds.
	ComponentWorkflowStepsQueueName = "component-workflow-steps"
)

// ComponentQueueIDs holds the queue IDs for a component.
type ComponentQueueIDs struct {
	DefaultQueueID       string `json:"default_queue_id"`
	WorkflowStepsQueueID string `json:"workflow_steps_queue_id"`
}

// EnsureComponentQueues creates all Temporal queue workflows for a component
// and returns the queue IDs. Safe to call multiple times — Create is idempotent.
func (h *Helpers) EnsureComponentQueues(ctx context.Context, componentID string) (*ComponentQueueIDs, error) {
	ownerType := plugins.TableName(h.db, app.Component{})

	defaultQueue, err := h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OwnerID:     componentID,
		OwnerType:   ownerType,
		Namespace:   "components",
		MaxInFlight: 1,
		MaxDepth:    50,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to ensure default queue for component %s: %w", componentID, err)
	}

	stepsQueue, err := h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OwnerID:     componentID,
		OwnerType:   ownerType,
		Namespace:   "components",
		Name:        ComponentWorkflowStepsQueueName,
		MaxInFlight: 10,
		MaxDepth:    50,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to ensure %s queue for component %s: %w", ComponentWorkflowStepsQueueName, componentID, err)
	}

	return &ComponentQueueIDs{
		DefaultQueueID:       defaultQueue.ID,
		WorkflowStepsQueueID: stepsQueue.ID,
	}, nil
}

// GetComponentQueueIDs looks up existing queue IDs for a component.
func (h *Helpers) GetComponentQueueIDs(ctx context.Context, componentID string) (*ComponentQueueIDs, error) {
	ownerType := plugins.TableName(h.db, app.Component{})

	defaultQueue, err := h.queueClient.GetQueueByOwner(ctx, componentID, ownerType)
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
