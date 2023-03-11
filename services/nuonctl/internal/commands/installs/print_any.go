package installs

import (
	"context"

	"github.com/powertoolsdev/mono/services/nuonctl/internal/proto"
)

func (c *commands) PrintRequest(ctx context.Context, key string) error {
	req, err := c.Workflows.GetInstallsRequest(ctx, key)
	if err != nil {
		return err
	}

	return proto.Print(req)
}

func (c *commands) PrintResponse(ctx context.Context, key string) error {
	req, err := c.Workflows.GetInstallsResponse(ctx, key)
	if err != nil {
		return err
	}

	return proto.Print(req)
}
