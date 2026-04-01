package pulumi

import (
	"time"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	ociarchive "github.com/nuonco/nuon/bins/runner/internal/pkg/oci/archive"
	pkgplantypes "github.com/nuonco/nuon/bins/runner/internal/pkg/plantypes"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/plugins/configs"
)

type handlerState struct {
	plan      *plantypes.DeployPlan
	pulumiCfg *models.AppPulumiComponentConfig

	auth *pkgplantypes.PlanAuth

	srcCfg  *configs.OCIRegistryRepository
	srcTag  string
	timeout time.Duration

	arch           ociarchive.Archive
	jobExecutionID string
	jobID          string
}
