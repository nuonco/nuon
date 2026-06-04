package pulumi

import (
	"time"

	pkgplantypes "github.com/nuonco/nuon/bins/runner/internal/pkg/plantypes"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/workspace"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	pulumiworkspace "github.com/nuonco/nuon/pkg/pulumi/workspace"
)

type handlerState struct {
	plan *plantypes.SandboxRunPlan

	auth *pkgplantypes.PlanAuth

	srcWorkspace workspace.Workspace
	workspace    *pulumiworkspace.Workspace
	timeout      time.Duration

	jobExecutionID string
	jobID          string

	outputs map[string]interface{}
}
