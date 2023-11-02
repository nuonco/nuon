package platform

import (
	"context"

	"github.com/hashicorp/waypoint-plugin-sdk/component"
	sdk "github.com/hashicorp/waypoint-plugin-sdk/proto/gen"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	jobv1 "github.com/powertoolsdev/mono/pkg/types/plugins/job/v1"
)

func (p *Platform) StatusFunc() interface{} {
	return p.status
}

func (p *Platform) status(
	ctx context.Context,
	ji *component.JobInfo,
	deploy *jobv1.Deployment,
	ui terminal.UI,
) (*sdk.StatusReport, error) {
	return &sdk.StatusReport{
		Health:        sdk.StatusReport_ALIVE,
		HealthMessage: "ok",
	}, nil
}
