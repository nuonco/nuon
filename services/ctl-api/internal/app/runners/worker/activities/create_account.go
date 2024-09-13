package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CreateAccountRequest struct {
	RunnerID string `validate:"required"`
	OrgID    string `validate:"required"`
}

// @temporal-gen activity
// @schedule-to-close-timeout 5s
func (a *Activities) CreateAccount(ctx context.Context, req CreateAccountRequest) (*app.Account, error) {
	acct, err := a.authzClient.CreateServiceAccount(ctx, req.RunnerID)
	if err != nil {
		return nil, fmt.Errorf("unable to create service account: %w", err)
	}

	a.authzClient.AddAccountOrgRole(ctx, app.RoleTypeRunner, req.OrgID, acct.ID)
	return acct, nil
}
