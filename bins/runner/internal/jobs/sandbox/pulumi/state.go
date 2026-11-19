package pulumi

import (
	"time"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	pulumiworkspace "github.com/nuonco/nuon/pkg/pulumi/workspace"
	ociarchive "github.com/nuonco/nuon/pkg/runner/oci/archive"
	pkgplantypes "github.com/nuonco/nuon/pkg/runner/plantypes"
	"github.com/nuonco/nuon/pkg/runner/workspace"
)

type handlerState struct {
	plan *plantypes.SandboxRunPlan

	auth *pkgplantypes.PlanAuth

	srcWorkspace workspace.Workspace
	ociArch      ociarchive.Archive
	workspace    *pulumiworkspace.Workspace
	timeout      time.Duration

	jobExecutionID string
	jobID          string

	outputs map[string]interface{}
}
