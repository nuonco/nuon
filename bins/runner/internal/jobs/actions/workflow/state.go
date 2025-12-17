package workflow

import (
	"github.com/nuonco/nuon/bins/runner/internal/pkg/workspace"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"

	"github.com/nuonco/nuon-runner-go/models"
)

type handlerState struct {
	// set during the fetch/validate phase
	workflowCfg *models.AppActionWorkflowConfig
	run         *models.AppInstallActionWorkflowRun
	plan        *plantypes.ActionWorkflowRunPlan

	// state that must be reset before each run
	workspace workspace.Workspace
}
