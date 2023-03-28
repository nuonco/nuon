package installs

import (
	"context"
	"fmt"
)

func (c *commands) Reprovision(ctx context.Context, installID string) error {
	req, err := c.Workflows.GetInstallProvisionRequest(ctx, installID)
	if err != nil {
		return fmt.Errorf("unable to get install provision request: %w", err)
	}

	err = c.Temporal.TriggerInstallProvision(ctx, req)
	if err != nil {
		return fmt.Errorf("unable to trigger provision: %w", err)
	}
	return nil
}
