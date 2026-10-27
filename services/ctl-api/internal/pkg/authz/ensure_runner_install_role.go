package authz

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func runnerInstallRole(orgID, installID string) *app.Role {
	return &app.Role{
		OrgID:       generics.NewNullString(orgID),
		RoleType:    app.RoleTypeRunnerInstall,
		Version:     app.RoleVersionV2,
		Title:       "Runner install",
		Description: "Scoped access to one install for a runner.",
		Policies: []app.Policy{{
			OrgID: generics.NewNullString(orgID),
			Name:  app.PolicyNameRunner,
			Permissions: pgtype.Hstore{
				permissions.APIPermissionKey:                                  generics.ToPtr(string(permissions.APIRunner)),
				permissions.Object(orgID, permissions.KindInstall, installID): permissions.PermissionAll.ToStrPtr(),
			},
		}},
	}
}

func (h *Client) EnsureRunnerInstallRole(ctx context.Context, orgID, installID, accountID string) error {
	if orgID == "" || installID == "" || accountID == "" {
		return fmt.Errorf("runner install role requires an org, install, and account")
	}
	ctx = cctx.SetAccountIDContext(ctx, accountID)
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing app.Role
		err := tx.Joins("JOIN account_roles ON account_roles.role_id = roles.id AND account_roles.deleted_at = 0").
			Where("account_roles.account_id = ?", accountID).
			Where(app.Role{OrgID: generics.NewNullString(orgID), RoleType: app.RoleTypeRunnerInstall}).
			First(&existing).Error
		if err == nil {
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("load runner install role: %w", err)
		}
		return ensurePrivateRole(tx, accountID, runnerInstallRole(orgID, installID))
	})
}
