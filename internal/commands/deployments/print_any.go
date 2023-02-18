package deployments

import (
	"context"

	"github.com/powertoolsdev/nuonctl/internal/proto"
)

func (c *commands) PrintRequest(ctx context.Context, key string) error {
	req, err := c.Workflows.GetDeploymentsRequest(ctx, key)
	if err != nil {
		return err
	}

	return proto.Print(req)
}

func (c *commands) PrintResponse(ctx context.Context, key string) error {
	req, err := c.Workflows.GetDeploymentsResponse(ctx, key)
	if err != nil {
		return err
	}

	return proto.Print(req)
}

func (c *commands) PrintPlan(ctx context.Context, key string) error {
	plan, err := c.Executors.GetDeploymentsPlan(ctx, key)
	if err != nil {
		return err
	}

	return proto.Print(plan)
}
