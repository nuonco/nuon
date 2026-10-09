package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
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

	var stack app.InstallStack
	if res := a.db.WithContext(ctx).
		Where(app.InstallStack{ID: req.InstallStackID}).
		First(&stack); res.Error != nil {
		return nil, generics.TemporalGormError(res.Error, "unable to load install stack: %w")
	}

	var accountID string
	if err := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		binding, err := account.New(account.Params{DB: tx}).EnsureManagedServiceAccount(ctx, account.ManagedServiceAccountRequest{
			OrgID: stack.OrgID, OwnerType: plugins.TableName(tx, app.InstallStack{}), OwnerID: stack.ID,
			Purpose: app.ManagedServiceAccountPurposeStack, InstanceKey: "default", Subject: stack.ID,
		})
		if err != nil {
			return fmt.Errorf("unable to ensure stack service account: %w", err)
		}
		accountID = binding.AccountID
		authzClient := authz.New(authz.Params{DB: tx})
		if err := authzClient.EnsureStackInstallRole(ctx, stack.OrgID, stack.InstallID, accountID); err != nil {
			return fmt.Errorf("unable to ensure stack service account role: %w", err)
		}
		return authzClient.RemoveAccountOrgRoleByType(ctx, app.RoleTypeOrgAdmin, stack.OrgID, accountID)
	}); err != nil {
		return nil, err
	}

	return &EnsureInstallStackServiceAccountResponse{AccountID: accountID}, nil
}
