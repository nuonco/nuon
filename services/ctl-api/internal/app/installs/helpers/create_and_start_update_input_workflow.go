package helpers

import (
	"context"
	"strconv"
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (h *Helpers) CreateAndStartInputUpdateWorkflow(
	ctx context.Context,
	installID string,
	changedInputs []string,
	changedInputValues string,
	role string,
	deployDependents bool,
) (*app.Workflow, error) {
	metadata := map[string]string{
		// NOTE(jm): this metadata field is not really designed to be used for anything serious, outside of
		// rendering things in the UI and other such things, which is why we are just using a string slice here,
		// maybe that will change at some point, but this metadata should not be abused.
		"inputs":            strings.Join(changedInputs, ","),
		"deploy_dependents": strconv.FormatBool(deployDependents),
	}
	if changedInputValues != "" {
		metadata[app.WorkflowMetadataKeyChangedInputValues] = changedInputValues
	}

	workflow, err := h.CreateWorkflowWithRole(
		ctx,
		installID,
		app.WorkflowTypeInputUpdate,
		metadata,
		false,
		role,
	)
	if err != nil {
		return nil, err
	}

	// Legacy evClient.Send calls removed — event loop system has been removed.
	// Callers in queue mode send their own queue signals (see update_install_input.go).

	return workflow, nil
}
