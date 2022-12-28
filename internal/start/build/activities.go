package build

import (
	"github.com/powertoolsdev/go-waypoint"
)

// NOTE(jm): we alias this type so it doesn't conflict
type waypointProvider = waypoint.Provider

type Activities struct {
	waypointProvider
	waypointApplicationUpserter
	waypointDeploymentJobPoller
	waypointDeploymentJobQueuer
	waypointDeploymentJobValidator
	artifactUploader
}

func NewActivities() *Activities {
	return &Activities{
		waypointProvider:               waypoint.NewProvider(),
		waypointApplicationUpserter:    &wpApplicationUpserter{},
		waypointDeploymentJobPoller:    &waypointDeploymentJobPollerImpl{},
		waypointDeploymentJobQueuer:    &waypointDeploymentJobQueuerImpl{},
		waypointDeploymentJobValidator: &waypointDeploymentJobValidatorImpl{},
		artifactUploader:               &artifactUploaderImpl{},
	}
}
