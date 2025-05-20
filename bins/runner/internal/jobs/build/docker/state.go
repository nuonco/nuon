package docker

import (
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/workspace"

	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
	"github.com/powertoolsdev/mono/pkg/plugins/configs"
)

type handlerState struct {
	// state for an individual run, that can not be reused
	plan           *plantypes.BuildPlan
	cfg            *plantypes.DockerBuildPlan
	regCfg         *configs.OCIRegistryRepository
	workspace      workspace.Workspace
	jobID          string
	jobExecutionID string
	resultTag      string
}
