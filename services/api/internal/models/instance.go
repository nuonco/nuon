package models

import (
	"github.com/powertoolsdev/mono/pkg/common/shortid/domains"
)

type Instance struct {
	ID string

	InstallID string
	Install   Install `faker:"-"`

	DeployID string
	Deploy   Deploy `faker:"-"`

	BuildID string
	Build   Build `faker:"-"`
}

func (i *Instance) NewID() error {
	if i.ID == "" {
		i.ID = domains.NewInstanceID()
	}
	return nil
}
