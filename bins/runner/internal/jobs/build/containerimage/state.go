package containerimage

import (
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/workspace"
	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
	"github.com/powertoolsdev/mono/pkg/plugins/configs"
)

type handlerState struct {
	// state for an individual run, that can not be reused
	plan      *plantypes.BuildPlan
	workspace workspace.Workspace

	jobID          string
	jobExecutionID string
	resultTag      string

	// the config can be one of the following:
	cfg    *plantypes.ContainerImagePullPlan
	regCfg *configs.OCIRegistryRepository
}
