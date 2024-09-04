package containerimage

import (
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/workspace"
	"github.com/powertoolsdev/mono/pkg/plugins/configs"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
)

type (
	WaypointConfig configs.App[configs.Build[configs.OCISyncBuild, configs.OCIRegistryRepository], configs.NoopDeploy]
)

type handlerState struct {
	// state for an individual run, that can not be reused
	plan      *planv1.Plan
	workspace workspace.Workspace

	jobID          string
	jobExecutionID string

	// the config can be one of the following:
	cfg       *configs.OCISyncBuild
	regCfg    *configs.OCIRegistryRepository
	resultTag string
}
