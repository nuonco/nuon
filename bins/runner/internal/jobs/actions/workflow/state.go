package workflow

import (
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	pkgplantypes "github.com/nuonco/nuon/pkg/runner/plantypes"
	"github.com/nuonco/nuon/pkg/runner/workspace"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

type handlerState struct {
	// set during the fetch/validate phase
	workflowCfg *models.AppActionWorkflowConfig
	run         *models.AppInstallActionWorkflowRun
	plan        *plantypes.ActionWorkflowRunPlan

	// state that must be reset before each run
	workspace workspace.Workspace

	auth *pkgplantypes.PlanAuth
}
