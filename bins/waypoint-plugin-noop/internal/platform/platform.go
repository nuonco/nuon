package platform

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/waypoint-plugin-sdk/component"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	noopv1 "github.com/powertoolsdev/mono/pkg/types/plugins/noop/v1"
)

func (p *Platform) DeployFunc() interface{} {
	return p.deploy
}

func (p *Platform) deploy(
	ctx context.Context,
	ji *component.JobInfo,
	ui terminal.UI,
	log hclog.Logger,
) (*noopv1.Deployment, error) {
	ui.Output("executing noop deploy...")
	return &noopv1.Deployment{}, nil
}
