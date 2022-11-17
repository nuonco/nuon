package project

import "github.com/powertoolsdev/go-waypoint"

type waypointProvider = waypoint.Provider

type Activities struct {
	waypointProvider
	waypointProjectCreator
}

func NewActivities() *Activities {
	return &Activities{
		waypointProjectCreator: &wpProjectCreator{},
	}
}
