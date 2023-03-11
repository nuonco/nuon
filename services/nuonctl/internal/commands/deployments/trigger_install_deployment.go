package deployments

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/common/shortid"
	deploymentsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/deployments/v1"
	"github.com/powertoolsdev/mono/services/nuonctl/internal/commands/deployments/presets"
)

func (c *commands) TriggerInstallPreset(ctx context.Context, installID, componentPreset string) error {
	req, err := c.Workflows.GetInstallProvisionRequest(ctx, installID)
	if err != nil {
		return fmt.Errorf("unable to get install provision request: %w", err)
	}

	presetComp, err := presets.New(c.v, componentPreset)
	if err != nil {
		return fmt.Errorf("unable to get preset: %w", err)
	}

	wkflowReq := &deploymentsv1.StartRequest{
		OrgId:        req.OrgId,
		AppId:        req.AppId,
		DeploymentId: shortid.New(),
		InstallIds:   []string{req.InstallId},
		Component:    presetComp,
		PlanOnly:     false,
	}
	if err := c.Temporal.TriggerDeploymentStart(ctx, wkflowReq); err != nil {
		return fmt.Errorf("unable to trigger deployment start: %w", err)
	}

	return nil
}
