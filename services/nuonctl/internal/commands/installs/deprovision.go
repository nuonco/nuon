package installs

import (
	"context"
	"fmt"

	installsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1"
)

func (c *commands) DeprovisionBulk(ctx context.Context, installIDs []string) error {
	for _, installID := range installIDs {
		if err := c.Deprovision(ctx, installID); err != nil {
			return fmt.Errorf("unable to deprovision install %s: %w", installID, err)
		}
	}

	return nil
}

func (c *commands) Deprovision(ctx context.Context, installID string) error {
	pReq, err := c.Workflows.GetInstallProvisionRequest(ctx, installID)
	if err != nil {
		return fmt.Errorf("unable to get install provision request: %w", err)
	}

	dReq := &installsv1.DeprovisionRequest{
		OrgId:           pReq.OrgId,
		AppId:           pReq.AppId,
		InstallId:       pReq.InstallId,
		AccountSettings: pReq.AccountSettings,
		SandboxSettings: pReq.SandboxSettings,
		PlanOnly:        false,
	}

	err = c.Temporal.TriggerInstallDeprovision(ctx, dReq)
	if err != nil {
		return fmt.Errorf("unable to trigger deprovision: %w", err)
	}
	return nil
}
