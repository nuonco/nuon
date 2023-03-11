package project

import "github.com/powertoolsdev/go-waypoint"

type Activities struct {
	waypoint.Provider
	waypointProjectCreator
	waypointWorkspaceUpserter
	waypointServerPinger
}

func NewActivities() *Activities {
	return &Activities{
		Provider:                  waypoint.NewProvider(),
		waypointProjectCreator:    &wpProjectCreator{},
		waypointWorkspaceUpserter: &wpWorkspaceUpserter{},
		waypointServerPinger:      &wpServerPinger{},
	}
}
