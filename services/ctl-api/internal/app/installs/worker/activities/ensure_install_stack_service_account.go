package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type EnsureInstallStackServiceAccountRequest struct {
	InstallStackID string `json:"install_stack_id" validate:"required"`
}

type EnsureInstallStackServiceAccountResponse struct {
	AccountID string `json:"account_id"`
}

// EnsureInstallStackServiceAccount reconciles the account an install stack
// authenticates as. Mints no token — those are created on demand from the dashboard.
//
// Keyed on the InstallStack, not the version, which regenerates on every config
// change. Convergent, so it also backfills older stacks.
//
// @temporal-gen-v2 activity
// @by-field InstallStackID
// @start-to-close-timeout 2m
func (a *Activities) EnsureInstallStackServiceAccount(
	ctx context.Context, req *EnsureInstallStackServiceAccountRequest,
) (*EnsureInstallStackServiceAccountResponse, error) {
	if err := a.v.StructCtx(ctx, req); err != nil {
		return nil, fmt.Errorf("unable to validate request: %w", err)
	}

	binding, err := a.acctClient.EnsureStack(ctx, req.InstallStackID)
	if err != nil {
		return nil, generics.TemporalGormError(err, "unable to ensure install stack account: %w")
	}

	return &EnsureInstallStackServiceAccountResponse{AccountID: binding.AccountID}, nil
}
