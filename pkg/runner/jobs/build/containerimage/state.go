package containerimage

import (
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/plugins/configs"
	"github.com/nuonco/nuon/pkg/runner/workspace"
)

type handlerState struct {
	plan      *plantypes.BuildPlan
	workspace workspace.Workspace

	jobID          string
	jobExecutionID string
	resultTag      string

	cfg    *plantypes.ContainerImagePullPlan
	regCfg *configs.OCIRegistryRepository
}
