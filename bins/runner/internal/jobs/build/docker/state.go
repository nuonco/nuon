package docker

import (
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/workspace"

	"github.com/powertoolsdev/mono/pkg/plugins/configs"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
)

// TODO(jm): this will eventually go away, once we remove the need for using a full waypoint app config
type InputConfig configs.App[configs.Build[configs.DockerBuild, configs.OCIRegistryRepository], configs.NoopDeploy]

type handlerState struct {
	// state for an individual run, that can not be reused
	plan           *planv1.Plan
	cfg            *configs.DockerBuild
	regCfg         *configs.OCIRegistryRepository
	workspace      workspace.Workspace
	jobID          string
	jobExecutionID string
	resultTag      string
}
