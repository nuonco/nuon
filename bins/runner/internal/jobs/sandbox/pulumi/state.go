package pulumi

import (
	"time"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	ociarchive "github.com/nuonco/nuon/bins/runner/internal/pkg/oci/archive"
	pkgplantypes "github.com/nuonco/nuon/bins/runner/internal/pkg/plantypes"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	pulumiworkspace "github.com/nuonco/nuon/pkg/pulumi/workspace"
)

type handlerState struct {
	plan      *plantypes.DeployPlan
	pulumiCfg *models.AppPulumiComponentConfig

	auth *pkgplantypes.PlanAuth

	arch      ociarchive.Archive
	workspace *pulumiworkspace.Workspace
	timeout   time.Duration

	jobExecutionID string
	jobID          string

	outputs map[string]interface{}
}
