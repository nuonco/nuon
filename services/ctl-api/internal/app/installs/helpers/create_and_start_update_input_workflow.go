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
	inputsOnly bool,
	planOnly bool,
	workflowType app.WorkflowType,
) (*app.Workflow, error) {
	metadata := map[string]string{
		"inputs":            strings.Join(changedInputs, ","),
		"deploy_dependents": strconv.FormatBool(deployDependents),
	}
	if inputsOnly {
		metadata[app.WorkflowMetadataKeyInputsOnly] = strconv.FormatBool(true)
	}
	if changedInputValues != "" {
		metadata[app.WorkflowMetadataKeyChangedInputValues] = changedInputValues
	}

	workflow, err := h.CreateWorkflowWithRole(
		ctx,
		installID,
		workflowType,
		metadata,
		planOnly,
		role,
	)
	if err != nil {
		return nil, err
	}

	return workflow, nil
}
