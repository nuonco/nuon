package helpers

import (
	"context"
	"strings"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
)

func (h *Helpers) CreateAndStartInputUpdateWorkflow(ctx context.Context, installID string, changedInputs []string) (*app.Workflow, error) {
	workflow, err := h.CreateWorkflow(ctx, installID, app.WorkflowTypeInputUpdate, map[string]string{
		// NOTE(jm): this metadata field is not really designed to be used for anything serious, outside of
		// rendering things in the UI and other such things, which is why we are just using a string slice here,
		// maybe that will change at some point, but this metadata should not be abused.
		"inputs": strings.Join(changedInputs, ","),
	}, app.StepErrorBehaviorAbort,
		false,
	)
	if err != nil {
		return nil, err
	}

	h.evClient.Send(ctx, installID, &signals.Signal{
		Type:              signals.OperationUpdated,
		InstallWorkflowID: workflow.ID,
	})
	h.evClient.Send(ctx, installID, &signals.Signal{
		Type:              signals.OperationExecuteFlow,
		InstallWorkflowID: workflow.ID,
	})

	return workflow, err
}
