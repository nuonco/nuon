package provision

import (
	"github.com/powertoolsdev/go-waypoint"
)

// NOTE: we alias this type so it doesn't conflict
type waypointProvider = waypoint.Provider

type Activities struct {
	waypointCfgGenerator
	waypointApplicationUpserter
	waypointProvider
	waypointDeploymentJobQueuer
	metadataUploader
}

func NewActivities() *Activities {
	return &Activities{
		waypointCfgGenerator:        &waypointCfgGeneratorImpl{},
		waypointApplicationUpserter: &wpApplicationUpserter{},
		waypointProvider:            waypoint.NewProvider(),
		waypointDeploymentJobQueuer: &waypointDeploymentJobQueuerImpl{},
		metadataUploader:            &metadataUploaderImpl{},
	}
}
