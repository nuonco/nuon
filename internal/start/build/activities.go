package build

import (
	"github.com/powertoolsdev/go-waypoint"
	workers "github.com/powertoolsdev/workers-deployments/internal"
)

// NOTE(jm): we alias this type so it doesn't conflict
type waypointProvider = waypoint.Provider

type Activities struct {
	waypointCfgGenerator
	waypointProvider
	waypointWorkspaceUpserter
	waypointApplicationUpserter
	waypointDeploymentJobQueuer
	waypointDeploymentJobValidator
	artifactUploader
}

func NewActivities(cfg workers.Config) *Activities {
	return &Activities{
		waypointCfgGenerator:           &waypointCfgGeneratorImpl{},
		waypointWorkspaceUpserter:      &wpWorkspaceUpserter{},
		waypointProvider:               waypoint.NewProvider(),
		waypointApplicationUpserter:    &wpApplicationUpserter{},
		waypointDeploymentJobQueuer:    &waypointDeploymentJobQueuerImpl{},
		waypointDeploymentJobValidator: &waypointDeploymentJobValidatorImpl{},
		artifactUploader:               &artifactUploaderImpl{},
	}
}
