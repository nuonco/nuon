package project

import "github.com/powertoolsdev/go-waypoint"

type Activities struct {
	waypoint.Provider
	waypointProjectCreator
	waypointWorkspaceUpserter
}

func NewActivities() *Activities {
	return &Activities{
		Provider:                  waypoint.NewProvider(),
		waypointProjectCreator:    &wpProjectCreator{},
		waypointWorkspaceUpserter: &wpWorkspaceUpserter{},
	}
}
