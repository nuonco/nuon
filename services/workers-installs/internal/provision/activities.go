package provision

import (
	"github.com/go-playground/validator/v10"
	"k8s.io/client-go/rest"

	"github.com/powertoolsdev/mono/pkg/sender"
	workers "github.com/powertoolsdev/mono/services/workers-installs/internal"
)

// ProvisionActivities is a type that wraps the set of provision activities that we'll be using to execute this
// workflow. It should only be a few activities, such as running terraform and installing the agent
type Activities struct {
	v      *validator.Validate
	sender sender.NotificationSender

	config *workers.Config

	// this is exposed for testing and should not otherwise be used
	Kubeconfig *rest.Config

	notifier
}

func NewActivities(v *validator.Validate, cfg *workers.Config, sender sender.NotificationSender) *Activities {
	return &Activities{
		v:        v,
		config:   cfg,
		sender:   sender,
		notifier: &notifierImpl{sender},
	}
}
