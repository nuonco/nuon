package platform

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/waypoint-plugin-sdk/component"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	terraformv1 "github.com/powertoolsdev/mono/pkg/types/plugins/terraform/v1"
)

func (p *Platform) DestroyFunc() interface{} {
	return p.destroy
}

func (p *Platform) destroy(
	ctx context.Context,
	ji *component.JobInfo,
	ui terminal.UI,
	log hclog.Logger,
) (*terraformv1.Deployment, error) {
	return p.execRun(ctx, runTypeDestroy, ji, ui, log)
}
