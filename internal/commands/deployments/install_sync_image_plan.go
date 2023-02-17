package deployments

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/nuonctl/internal/proto"
)

func (c *commands) InstallSyncImagePlan(ctx context.Context, installID, componentPreset string) error {
	req, err := c.installPresetRequest(ctx, installID, componentPreset)
	if err != nil {
		return fmt.Errorf("unable to get install preset request: %w", err)
	}

	_, err = c.Temporal.ExecDeploymentStart(ctx, req)
	if err != nil {
		return fmt.Errorf("unable to execute deployment start: %w", err)
	}

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

	syncImagePlan, err := c.Executors.GetPlan(ctx, instance.ImageSyncPlan)
	if err != nil {
		return fmt.Errorf("unable to get sync image plan: %w", err)
	}

	return proto.Print(syncImagePlan)
}
