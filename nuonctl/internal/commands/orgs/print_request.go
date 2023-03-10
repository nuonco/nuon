package orgs

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/nuonctl/internal/proto"
)

func (c *commands) PrintProvisionRequest(ctx context.Context, orgID string) error {
	req, err := c.Workflows.GetOrgProvisionRequest(ctx, orgID)
	if err != nil {
		return fmt.Errorf("unable to get org provision request: %w", err)
	}

	return proto.Print(req)
}

func (c *commands) PrintProvisionResponse(ctx context.Context, orgID string) error {
	resp, err := c.Workflows.GetOrgProvisionResponse(ctx, orgID)
	if err != nil {
		return fmt.Errorf("unable to get org provision response: %w", err)
	}

	return proto.Print(resp)
}
