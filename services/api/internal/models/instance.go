package models

import (
	"github.com/powertoolsdev/mono/pkg/common/shortid/domains"
)

type Instance struct {
	ID string

	InstallID string
	Install   Install `faker:"-"`

	BuildID string
	Build   Build `faker:"-"`

	ComponentID string
	Component   Component `faker:"-"`
}

func (i *Instance) NewID() {
	if i.ID == "" {
		i.ID = domains.NewInstanceID()
	}
}
