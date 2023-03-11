package orgs

import (
	"context"
	"fmt"

	orgsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1"
)

func (c *commands) Provision(ctx context.Context, orgID string) error {
	req := &orgsv1.SignupRequest{OrgId: orgID, Region: "us-west-2"}

	if err := c.Temporal.TriggerOrgSignup(ctx, req); err != nil {
		return fmt.Errorf("unable to trigger signup: %w", err)
	}

	return nil
}
