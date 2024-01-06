package platform

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/waypoint-plugin-sdk/component"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	"github.com/powertoolsdev/mono/pkg/plugins/configs"
	terraformv1 "github.com/powertoolsdev/mono/pkg/types/plugins/terraform/v1"
)

func (p *Platform) DestroyFunc() interface{} {
	return p.destroy
}

func (p *Platform) destroy(
	ctx context.Context,
	ji *component.JobInfo,
	src *component.Source,
	ui terminal.UI,
	log hclog.Logger,
) (*terraformv1.Deployment, error) {
	if p.Cfg.RunType != configs.TerraformDeployRunTypeDestroy {
		return nil, fmt.Errorf("invalid run type for destroy operation")
	}

	return p.execRun(ctx, ji, src, ui, log)
}
