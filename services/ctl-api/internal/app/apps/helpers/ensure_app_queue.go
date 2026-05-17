package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

const (
	AppWorkflowsQueueName          = "app-workflows"
	AppSignalsQueueName            = "app-signals"
	AppWorkflowStepGroupsQueueName = "app-workflow-step-groups"
	AppWorkflowStepsQueueName      = "app-workflow-steps"
	AppGenerateStepsQueueName      = "app-generate-steps"
)

// EnsureAppQueue creates all Temporal queue workflows needed for an app to
// execute workflows through the shared flow infrastructure.
// Safe to call multiple times — queueClient.Create is idempotent.
func (h *Helpers) EnsureAppQueue(ctx context.Context, appID string) error {
	ownerType := plugins.TableName(h.db, app.App{})

	// app-workflows queue — orchestrates workflow execution (executeflow.Signal)
	if _, err := h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OwnerID:     appID,
		OwnerType:   ownerType,
		Namespace:   "apps",
		Name:        AppWorkflowsQueueName,
		MaxInFlight: 2,
		MaxDepth:    50,
	}); err != nil {
		return fmt.Errorf("unable to ensure app-workflows queue for app %s: %w", appID, err)
	}

	// app-signals queue — handles individual signal execution (generate-workflow-steps, component builds)
	if _, err := h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OwnerID:     appID,
		OwnerType:   ownerType,
		Namespace:   "apps",
		Name:        AppSignalsQueueName,
		MaxInFlight: 20,
		MaxDepth:    50,
	}); err != nil {
		return fmt.Errorf("unable to ensure app-signals queue for app %s: %w", appID, err)
	}

	// app-workflow-step-groups queue — executes step groups
	if _, err := h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OwnerID:     appID,
		OwnerType:   ownerType,
		Namespace:   "apps",
		Name:        AppWorkflowStepGroupsQueueName,
		MaxInFlight: 10,
		MaxDepth:    50,
	}); err != nil {
		return fmt.Errorf("unable to ensure app-workflow-step-groups queue for app %s: %w", appID, err)
	}

	// app-workflow-steps queue — executes individual workflow steps
	if _, err := h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OwnerID:     appID,
		OwnerType:   ownerType,
		Namespace:   "apps",
		Name:        AppWorkflowStepsQueueName,
		MaxInFlight: 10,
		MaxDepth:    50,
	}); err != nil {
		return fmt.Errorf("unable to ensure app-workflow-steps queue for app %s: %w", appID, err)
	}

	// app-generate-steps queue — handles generate-steps signals
	if _, err := h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OwnerID:     appID,
		OwnerType:   ownerType,
		Namespace:   "apps",
		Name:        AppGenerateStepsQueueName,
		MaxInFlight: 10,
		MaxDepth:    50,
	}); err != nil {
		return fmt.Errorf("unable to ensure app-generate-steps queue for app %s: %w", appID, err)
	}

	return nil
}
