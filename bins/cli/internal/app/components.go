package app

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/ui"
)

func (c *commands) ListComponents(ctx context.Context) error {
	if err := c.ensureAppID(); err != nil {
		return err
	}

	cmpsResp, err := c.apiClient.GetComponents(ctx, c.appID)
	if err != nil {
		return fmt.Errorf("unable to get apps: %w", err)
	}

	for _, cmp := range cmpsResp {
		ui.Line(ctx, "%s - %s", cmp.Id, cmp.Name)
	}
	return nil
}
