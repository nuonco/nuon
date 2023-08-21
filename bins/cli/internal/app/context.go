package app

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/ui"
)

func (c *commands) GetContext(ctx context.Context) error {
	if err := c.ensureOrgID(); err != nil {
		return err
	}

	orgResp, err := c.client.GetOrg(ctx)
	if err != nil {
		return fmt.Errorf("unable to get org: %w", err)
	}

	statusColor := ui.GetStatusColor(orgResp.Status)

	ui.Line(ctx, "%s%s %s- %s - %s", statusColor, orgResp.Status, ui.ColorReset, orgResp.ID, orgResp.Name)
	return nil
}
