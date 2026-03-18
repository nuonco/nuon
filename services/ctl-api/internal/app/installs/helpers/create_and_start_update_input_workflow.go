package helpers

import (
	"context"
	"fmt"
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	qsignal "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// CreateInputUpdateWorkflow creates an input update workflow without sending signals.
// Callers are responsible for sending the appropriate signals after calling this method.
func (h *Helpers) CreateInputUpdateWorkflow(ctx context.Context, installID string, changedInputs []string) (*app.Workflow, error) {
	return h.CreateWorkflow(ctx, installID, app.WorkflowTypeInputUpdate, map[string]string{
		// NOTE(jm): this metadata field is not really designed to be used for anything serious, outside of
		// rendering things in the UI and other such things, which is why we are just using a string slice here,
		// maybe that will change at some point, but this metadata should not be abused.
		"inputs": strings.Join(changedInputs, ","),
	},
		false,
	)
}

// CreateAndStartInputUpdateWorkflow creates an input update workflow and sends signals.
// It accepts v2 signal constructors so callers from packages that can import v2 signal types
// can provide them, while callers that can't (due to import cycles) pass nil to use legacy only.
func (h *Helpers) CreateAndStartInputUpdateWorkflow(ctx context.Context, installID string, changedInputs []string, v2Signals []qsignal.Signal) (*app.Workflow, error) {
	workflow, err := h.CreateInputUpdateWorkflow(ctx, installID, changedInputs)
	if err != nil {
		return nil, err
	}

	if err := h.SendInstallSignals(ctx, installID, workflow.ID, v2Signals); err != nil {
		return nil, err
	}

	return workflow, nil
}

// SendInstallSignals sends v2 queue signals if queues are enabled (and v2Signals are provided),
// otherwise falls back to legacy event loop signals for updated + execute-flow.
func (h *Helpers) SendInstallSignals(ctx context.Context, installID, workflowID string, v2Signals []qsignal.Signal) error {
	useQueues, err := h.featuresClient.AllFeaturesEnabled(ctx, app.OrgFeatureAppBranches, app.OrgFeatureQueues)
	if err != nil {
		return fmt.Errorf("checking features: %w", err)
	}

	if useQueues && len(v2Signals) > 0 {
		var queue app.Queue
		if res := h.db.WithContext(ctx).Where("owner_id = ?", installID).First(&queue); res.Error != nil {
			return fmt.Errorf("unable to get install queue: %w", res.Error)
		}

		for _, sig := range v2Signals {
			if _, err := h.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
				QueueID: queue.ID,
				Signal:  sig,
			}); err != nil {
				return fmt.Errorf("enqueue signal: %w", err)
			}
		}
	} else {
		h.evClient.Send(ctx, installID, &signals.Signal{
			Type:              signals.OperationUpdated,
			InstallWorkflowID: workflowID,
		})
		h.evClient.Send(ctx, installID, &signals.Signal{
			Type:              signals.OperationExecuteFlow,
			InstallWorkflowID: workflowID,
		})
	}

	return nil
}
