package platform

import (
	"context"

	"github.com/hashicorp/go-hclog"
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
	ui terminal.UI,
	log hclog.Logger,
) (*terraformv1.Deployment, error) {
	runTyp := runTypeApply
	if p.Cfg.PlanOnly {
		runTyp = runTypePlan
	}

	return p.execRun(ctx, runTyp, ji, ui, log)
}
