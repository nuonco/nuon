package terraform

import (
	"time"

	"github.com/nuonco/nuon-runner-go/models"

	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/workspace"
	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
	terraformworkspace "github.com/powertoolsdev/mono/pkg/terraform/workspace"
)

const (
	defaultFileType string = "file/terraform"
)

type handlerState struct {
	workspace workspace.Workspace

	timeout time.Duration

	// fields set by the plugin execution
	jobExecutionID string
	jobID          string
	tfWorkspace    terraformworkspace.Workspace

	plan       *plantypes.SandboxRunPlan
	appCfg     *models.AppAppConfig
	sandboxCfg *models.AppAppSandboxConfig

	// Legacy
	// set during the fetch/validate phase
	// plan    *planv1.Plan
	// cfg     *configs.SandboxTerraform
}
