package pulumi

import (
	"time"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/plugins/configs"
	pulumiworkspace "github.com/nuonco/nuon/pkg/pulumi/workspace"
	ociarchive "github.com/nuonco/nuon/pkg/runner/oci/archive"
	pkgplantypes "github.com/nuonco/nuon/pkg/runner/plantypes"
)

type handlerState struct {
	plan      *plantypes.DeployPlan
	pulumiCfg *models.AppPulumiComponentConfig

	auth *pkgplantypes.PlanAuth

	srcCfg  *configs.OCIRegistryRepository
	srcTag  string
	timeout time.Duration

	arch           ociarchive.Archive
	workspace      *pulumiworkspace.Workspace
	jobExecutionID string
	jobID          string

	// outputs captured from pulumi up, returned via Outputs()
	outputs map[string]interface{}
}
