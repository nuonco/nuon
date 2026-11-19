package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type CreateAccountRequest struct {
	RunnerID string `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) CreateAccount(ctx context.Context, req CreateAccountRequest) (*app.Account, error) {
	return a.acctClient.EnsureRunner(ctx, req.RunnerID)
}
