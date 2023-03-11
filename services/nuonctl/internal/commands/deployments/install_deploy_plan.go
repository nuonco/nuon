package deployments

import (
	"context"
	"fmt"
	"time"

	"github.com/powertoolsdev/mono/services/nuonctl/internal/proto"
)

func (c *commands) InstallDeployPlan(ctx context.Context, installID, componentPreset string, planOnly bool) error {
	req, err := c.installPresetRequest(ctx, installID, componentPreset)
	if err != nil {
		return fmt.Errorf("unable to get install preset request: %w", err)
	}
	req.PlanOnly = planOnly

	_, err = c.Temporal.ExecDeploymentStart(ctx, req)
	if err != nil {
		return fmt.Errorf("unable to execute deployment start: %w", err)
	}

	// NOTE(jm): we wait an additional 5 seconds before checking instance as deployments currently don't wait for
	// instances to finish
	//nolint:all
	time.Sleep(time.Second * 5)
	resp, err := c.Workflows.GetInstanceProvisionResponse(ctx,
		req.OrgId,
		req.AppId,
		req.Component.Id,
		req.DeploymentId,
		req.InstallIds[0],
	)
	if err != nil {
		return fmt.Errorf("unable to get instance provision response: %w", err)
	}

	instance := resp.Response.GetInstanceProvision()
	if instance == nil {
		return fmt.Errorf("invalid response")
	}

	if instance.DeployPlan == nil {
		return fmt.Errorf("no deploy plan set")
	}

	deployPlan, err := c.Executors.GetPlan(ctx, instance.DeployPlan)
	if err != nil {
		return fmt.Errorf("unable to get deploy plan: %w", err)
	}

	return proto.Print(deployPlan)
}
