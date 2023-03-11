package installs

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/nuonctl/internal/proto"
)

func (c *commands) PrintProvisionRequest(ctx context.Context, installID string) error {
	req, err := c.Workflows.GetInstallProvisionRequest(ctx, installID)
	if err != nil {
		return fmt.Errorf("unable to get install provision request: %w", err)
	}

	return proto.Print(req)
}

func (c *commands) PrintProvisionResponse(ctx context.Context, installID string) error {
	return nil
}
