package instances

import workers "github.com/powertoolsdev/workers-deployments/internal"

type Activities struct {
	provisioner
}

func NewActivities(cfg workers.Config) *Activities {
	return &Activities{
		provisioner: &instanceProvisioner{
			TemporalHost:      cfg.TemporalHost,
			TemporalNamespace: cfg.TemporalNamespace,
		},
	}
}
