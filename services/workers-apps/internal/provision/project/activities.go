package project

import (
	"github.com/go-playground/validator/v10"
)

type Activities struct {
	v *validator.Validate

	waypointProjectCreator
	waypointWorkspaceUpserter
	waypointServerPinger
}

func NewActivities(v *validator.Validate) *Activities {
	return &Activities{
		v: v,

		waypointProjectCreator:    &wpProjectCreator{},
		waypointWorkspaceUpserter: &wpWorkspaceUpserter{},
		waypointServerPinger:      &wpServerPinger{},
	}
}
