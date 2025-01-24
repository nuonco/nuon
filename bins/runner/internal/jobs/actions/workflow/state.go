package workflow

import (
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/workspace"
	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"

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
