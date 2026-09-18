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
	AppWorkflowsQueueName          = queuenames.AppWorkflowsQueueName
	AppSignalsQueueName            = queuenames.AppSignalsQueueName
	AppWorkflowStepGroupsQueueName = queuenames.AppWorkflowStepGroupsQueueName
	AppWorkflowStepsQueueName      = queuenames.AppWorkflowStepsQueueName
	AppGenerateStepsQueueName      = queuenames.AppGenerateStepsQueueName
	AppInstallSyncsQueueName       = queuenames.AppInstallSyncsQueueName
)

// ensureAppQueueByName creates the named app queue at its registered capacity.
// Safe to call multiple times — queueClient.Create is idempotent and reconciles
// capacity drift against the registry.
func (h *Helpers) ensureAppQueueByName(ctx context.Context, appID, name string, skipRestartHint bool) (*app.Queue, error) {
	spec, ok := queuenames.SpecByName(queuenames.OwnerApps, name)
	if !ok {
		return nil, fmt.Errorf("app queue %q is not registered", name)
	}
	q, err := h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OwnerID:         appID,
		OwnerType:       plugins.TableName(h.db, app.App{}),
		Namespace:       "apps",
		Name:            spec.Name,
		MaxInFlight:     spec.MaxInFlight,
		MaxDepth:        spec.MaxDepth,
		SkipRestartHint: skipRestartHint,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to ensure %s queue for app %s: %w", spec.Name, appID, err)
	}
	return q, nil
}

func (h *Helpers) EnsureAppTriggerQueue(ctx context.Context, appID string) (*app.Queue, error) {
	return h.ensureAppQueueByName(ctx, appID, queuenames.AppTriggersQueueName, true)
}

// EnsureAppQueue creates all Temporal queue workflows needed for an app to
// execute workflows through the shared flow infrastructure.
func (h *Helpers) EnsureAppQueue(ctx context.Context, appID string) error {
	specs, ok := queuenames.Specs(queuenames.OwnerApps)
	if !ok {
		return fmt.Errorf("app queue specs are not registered")
	}
	for _, spec := range specs {
		if _, err := h.ensureAppQueueByName(ctx, appID, spec.Name, false); err != nil {
			return err
		}
	}

	return nil
}
