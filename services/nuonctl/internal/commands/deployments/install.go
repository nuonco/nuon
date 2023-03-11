package deployments

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/common/shortid"
	deploymentsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/deployments/v1"
	"github.com/powertoolsdev/mono/services/nuonctl/internal/commands/deployments/presets"
)

// this file contains shared tooling for emitting deployments and observing a single install

// installPresetRequest returns a request for a preset, for the specified install
func (c *commands) installPresetRequest(ctx context.Context, installID string, componentPreset string) (*deploymentsv1.StartRequest, error) {
	req, err := c.Workflows.GetInstallProvisionRequest(ctx, installID)
	if err != nil {
		return nil, fmt.Errorf("unable to get install provision request: %w", err)
	}

	presetComp, err := presets.New(c.v, componentPreset)
	if err != nil {
		return nil, fmt.Errorf("unable to get preset: %w", err)
	}

	return &deploymentsv1.StartRequest{
		OrgId:        req.OrgId,
		AppId:        req.AppId,
		DeploymentId: shortid.New(),
		InstallIds:   []string{installID},
		Component:    presetComp,
	}, nil
}
