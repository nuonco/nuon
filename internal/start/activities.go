package start

import (
	workers "github.com/powertoolsdev/workers-deployments/internal"
)

type Activities struct {
	starter starter

	provisioner
}

func NewActivities(cfg workers.Config) *Activities {
	return &Activities{
		starter: &starterImpl{},
		provisioner: &instanceProvisioner{
			TemporalHost:      cfg.TemporalHost,
			TemporalNamespace: cfg.TemporalNamespace,
		},
	}
}
