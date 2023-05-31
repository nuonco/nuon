package platform

import (
	"context"

	"github.com/hashicorp/waypoint-plugin-sdk/component"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	terraformv1 "github.com/powertoolsdev/mono/pkg/types/plugins/terraform/v1"
)

func (p *Platform) DeployFunc() interface{} {
	return p.deploy
}

func (p *Platform) deploy(
	ctx context.Context,
	ji *component.JobInfo,
	// artifact is the output of the push function, for the registry
	artifact *terraformv1.Artifact,
	ui terminal.UI,
) (*terraformv1.Deployment, error) {
	//
	return nil, nil
}
