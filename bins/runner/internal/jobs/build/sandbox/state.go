package sandbox

import (
	ociarchive "github.com/nuonco/nuon/bins/runner/internal/pkg/oci/archive"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/workspace"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/plugins/configs"
)

const defaultFileType = "file/terraform"

type handlerState struct {
	plan      *plantypes.BuildPlan
	cfg       *plantypes.TerraformBuildPlan
	workspace workspace.Workspace
	arch      ociarchive.Archive

	// regCfg is optional — if nil the build validates source but skips OCI push.
	regCfg    *configs.OCIRegistryRepository
	resultTag string

	jobID          string
	jobExecutionID string
}
