package deployments

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/nuonctl/internal/proto"
)

func (c *commands) InstallBuild(ctx context.Context, installID, componentPreset string, planOnly bool) error {
	req, err := c.installPresetRequest(ctx, installID, componentPreset)
	if err != nil {
		return fmt.Errorf("unable to get install preset request: %w", err)
	}
	req.BuildOnly = true
	req.PlanOnly = planOnly

	resp, err := c.Temporal.ExecDeploymentStart(ctx, req)
	if err != nil {
		return fmt.Errorf("unable to execute deployment start: %w", err)
	}

	plan, err := c.Executors.GetPlan(ctx, resp.PlanRef)
	if err != nil {
		return fmt.Errorf("unable to get plan: %w", err)
	}

	return proto.Print(plan)
}
