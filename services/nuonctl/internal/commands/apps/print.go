package apps

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/nuonctl/internal/proto"
)

func (c *commands) PrintProvisionRequest(ctx context.Context, orgID, appID string) error {
	req, err := c.Workflows.GetAppProvisionRequest(ctx, orgID, appID)
	if err != nil {
		return fmt.Errorf("unable to get app provision request: %w", err)
	}

	return proto.Print(req)
}

func (c *commands) PrintProvisionResponse(ctx context.Context, orgID, appID string) error {
	resp, err := c.Workflows.GetAppProvisionResponse(ctx, orgID, appID)
	if err != nil {
		return fmt.Errorf("unable to get app provision response: %w", err)
	}

	return proto.Print(resp)
}
