package workflow

import (
	pkgplantypes "github.com/nuonco/nuon/bins/runner/internal/pkg/plantypes"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/runner/workspace"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

type handlerState struct {
	workflowCfg *models.AppActionWorkflowConfig
	run         *models.AppInstallActionWorkflowRun
	plan        *plantypes.ActionWorkflowRunPlan

	workspace workspace.Workspace

	auth *pkgplantypes.PlanAuth
}
