package docker

import (
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/workspace"

	"github.com/powertoolsdev/mono/pkg/plugins/configs"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
)

type (
	Registry configs.Registry[configs.OCIRegistryRepository]
	Build    configs.Build[configs.DockerBuild, Registry]
	Deploy   configs.Deploy[configs.NoopDeploy]

	WaypointConfig configs.Apps[Build, Deploy]
)

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
