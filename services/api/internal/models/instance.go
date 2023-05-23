package models

import (
	"fmt"

	"github.com/powertoolsdev/mono/pkg/common/shortid"
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
		id, err := shortid.NewNanoID("ins")
		if err != nil {
			return fmt.Errorf("unable to make nanoid for instance: %w", err)
		}
		i.ID = id
	}
	return nil
}
