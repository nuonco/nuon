package app

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/ui"
)

func (c *commands) ListInstalls(ctx context.Context) error {
	if err := c.ensureAppID(); err != nil {
		return err
	}

	installsResp, err := c.apiClient.GetInstalls(ctx, c.appID)
	if err != nil {
		return fmt.Errorf("unable to get installs: %w", err)
	}

	for _, inst := range installsResp {
		ui.Line(ctx, "%s - %s (%s)", inst.Id, inst.Name, inst.GetSettings())
	}
	return nil
}
