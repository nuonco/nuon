package models

import (
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type Instance struct {
	Model

	InstallID string
	Install   Install `faker:"-"`

	ComponentID string
	Component   Component `faker:"-"`

	Deploys []*Deploy `faker:"-"`
}

func (i *Instance) NewID() {
	if i.ID == "" {
		i.ID = domains.NewInstanceID()
	}
}
