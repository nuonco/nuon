package general

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/common/shortid"
	canaryv1 "github.com/powertoolsdev/mono/pkg/types/workflows/canary/v1"
)

func (c *commands) ProvisionCanary(ctx context.Context) error {
	req := &canaryv1.ProvisionRequest{
		CanaryId: shortid.New(),
		Tags: map[string]string{
			"triggered-by": "nuonctl",
		},
	}

	if err := c.Temporal.TriggerCanaryProvision(ctx, req); err != nil {
		return fmt.Errorf("unable to provision canary: %w", err)
	}

	return nil
}

func (c *commands) DeprovisionCanary(ctx context.Context, canaryID string) error {
	req := &canaryv1.ProvisionRequest{
		CanaryId: canaryID,
		Tags: map[string]string{
			"triggered-by": "nuonctl",
		},
	}

	if err := c.Temporal.TriggerCanaryProvision(ctx, req); err != nil {
		return fmt.Errorf("unable to deprovision canary: %w", err)
	}

	return nil
}
