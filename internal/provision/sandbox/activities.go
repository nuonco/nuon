package sandbox

import (
	"k8s.io/client-go/rest"

	workers "github.com/powertoolsdev/workers-installs/internal"
)

// ProvisionActivities is a type that wraps the set of provision activities that we'll be using to execute this
// workflow. It should only be a few activities, such as running terraform and installing the agent
type Activities struct {
	terraformApplyer terraformApplyer

	config workers.Config

	// this is exposed for testing and should not otherwise be used
	Kubeconfig *rest.Config
}

func NewActivities(cfg workers.Config) *Activities {
	return &Activities{
		terraformApplyer: &tfApplyer{},
		config:           cfg,
	}
}
