package platform

import (
	"context"
	"fmt"

	"github.com/hashicorp/waypoint-plugin-sdk/component"
	sdk "github.com/hashicorp/waypoint-plugin-sdk/proto/gen"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	"github.com/powertoolsdev/mono/pkg/terraform/run"
	terraformv1 "github.com/powertoolsdev/mono/pkg/types/plugins/terraform/v1"
)

func (p *Platform) StatusFunc() interface{} {
	return p.status
}

func (p *Platform) status(
	ctx context.Context,
	ji *component.JobInfo,
	deploy *terraformv1.Deployment,
	ui terminal.UI,
) (*sdk.StatusReport, error) {
	tfRun, err := run.New(p.v, run.WithWorkspace(p.Workspace))
	if err != nil {
		return nil, fmt.Errorf("unable to create run: %w", err)
	}

	if err := tfRun.Plan(ctx); err != nil {
		return nil, fmt.Errorf("unable to get status: %w", err)
	}

	return nil, nil
}
