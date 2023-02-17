package deployments

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/powertoolsdev/go-common/shortid"
	"github.com/powertoolsdev/nuonctl/internal/commands/deployments/presets"
	deploymentsv1 "github.com/powertoolsdev/protos/workflows/generated/types/deployments/v1"
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

	deploymentID := uuid.NewString()
	ids, err := shortid.ToUUIDs(req.OrgId, req.AppId, req.InstallId)
	if err != nil {
		return fmt.Errorf("invalid install ids: %w", err)
	}

	wkflowReq := &deploymentsv1.StartRequest{
		OrgId:        ids[0].String(),
		AppId:        ids[1].String(),
		DeploymentId: deploymentID,
		InstallIds:   []string{ids[2].String()},
		Component:    presetComp,
		PlanOnly:     false,
	}
	if err := c.Temporal.TriggerDeploymentStart(ctx, wkflowReq); err != nil {
		return fmt.Errorf("unable to trigger deployment start: %w", err)
	}

	return nil
}
