package project

import "github.com/powertoolsdev/go-waypoint"

type waypointProvider = waypoint.Provider

type Activities struct {
	waypointProvider
	waypointProjectCreator
	waypointWorkspaceUpserter
}

func NewActivities() *Activities {
	return &Activities{
		waypointProvider:          waypoint.NewProvider(),
		waypointProjectCreator:    &wpProjectCreator{},
		waypointWorkspaceUpserter: &wpWorkspaceUpserter{},
	}
}
