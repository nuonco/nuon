package runner

import (
	"github.com/powertoolsdev/go-helm"
	"github.com/powertoolsdev/go-waypoint"
	workers "github.com/powertoolsdev/workers-orgs/internal"
	"k8s.io/client-go/rest"
)

// NOTE(jm): we alias this type here so that it embeds as WaypointProvider, but allows us to have waypoint.Provider as
// the actual package
type WaypointProvider = waypoint.Provider

type Activities struct {
	waypointProvider WaypointProvider
	helmInstaller    installer
	waypointServerCookieGetter
	waypointRunnerAdopter
	roleBindingCreator
	odrIAMRoleCreator
	odrIAMPolicyCreator
	iamRoleAssumer

	config     workers.Config
	Kubeconfig *rest.Config
}

func NewActivities(cfg workers.Config) *Activities {
	return &Activities{
		waypointProvider:           waypoint.NewProvider(),
		waypointServerCookieGetter: &wpServerCookieGetter{},
		waypointRunnerAdopter:      &wpRunnerAdopter{},
		config:                     cfg,
		helmInstaller:              helm.NewInstaller(),
		roleBindingCreator:         &roleBindingCreatorImpl{},
		odrIAMRoleCreator:          &odrIAMRoleCreatorImpl{},
		odrIAMPolicyCreator:        &odrIAMPolicyCreatorImpl{},
		iamRoleAssumer:             &iamRoleAssumerImpl{},
	}
}
