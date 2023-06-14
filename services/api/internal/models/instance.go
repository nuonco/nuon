package models

import (
	"github.com/powertoolsdev/mono/pkg/common/shortid/domains"
)

type Instance struct {
	Model

	InstallID string
	Install   Install `faker:"-"`

	ComponentID string
	Component   Component `faker:"-"`
}

func (i *Instance) NewID() {
	if i.ID == "" {
		i.ID = domains.NewInstanceID()
	}
}
