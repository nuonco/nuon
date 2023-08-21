package app

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/ui"
)

func (c *commands) LegacyListApps(ctx context.Context) error {
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

func (c *commands) ListApps(ctx context.Context) error {
	if err := c.ensureOrgID(); err != nil {
		return err
	}

	appsRes, err := c.client.GetApps(ctx)
	if err != nil {
		return fmt.Errorf("unable to get apps %w", err)
	}

	for _, app := range appsRes {
		ui.Line(ctx, "%s - %s", app.ID, app.Name)
	}

	return nil
}

func (c *commands) GetApp(ctx context.Context) error {
	if err := c.ensureAppID(); err != nil {
		return err
	}

	appRes, err := c.client.GetApp(ctx, c.appID)
	if err != nil {
		return fmt.Errorf("unable to get app %w", err)
	}

	ui.Line(ctx, "%s - %s", appRes.ID, appRes.Name)

	return nil
}
