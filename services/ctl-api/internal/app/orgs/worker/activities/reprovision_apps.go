package activities

import (
	"context"
	"fmt"
)

type ReprovisionAppsRequest struct {
	OrgID string `validate:"required"`
}

func (a *Activities) ReprovisionApps(ctx context.Context, req ReprovisionAppsRequest) error {
	org, err := a.getOrg(ctx, req.OrgID)
	if err != nil {
		return fmt.Errorf("unable to get org: %w", err)
	}

	for _, app := range org.Apps {
		a.appHooks.Reprovision(ctx, app.ID)
	}

	return nil
}
