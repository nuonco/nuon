package app

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/ui"
)

func (c *commands) ListApps(ctx context.Context) error {
	if err := c.ensureOrgID(); err != nil {
		return err
	}

	appsResp, err := c.apiClient.GetApps(ctx, c.orgID)
	if err != nil {
		return fmt.Errorf("unable to get apps: %w", err)
	}

	for _, app := range appsResp {
		ui.Line(ctx, "%s - %s", app.Id, app.Name)
	}
	return nil
}
