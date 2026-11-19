package authz

import (
	"context"
	"fmt"
	"reflect"

	"github.com/jackc/pgx/v5/pgtype"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func (h *Client) EnsureRunnerRole(ctx context.Context, orgID, installID, accountID string) error {
	if orgID == "" || accountID == "" {
		return fmt.Errorf("runner role requires an org and account")
	}
	ctx = cctx.SetAccountIDContext(ctx, accountID)
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		want := &app.Role{
			OrgID: generics.NewNullString(orgID), RoleType: app.RoleTypeRunnerIdentity, Version: app.RoleVersionV2,
			Title: "Runner", Description: "Permissions for one runner identity on the runner API.",
			Policies: []app.Policy{orgPolicy(orgID, app.PolicyNameRunner, permissions.PermissionAll, permissions.APIRunner)},
		}
		if err := New(Params{DB: tx}).EnsureManagedServiceAccountRole(ctx, accountID, want); err != nil {
			return err
		}
		var existing []app.Role
		if err := tx.Preload("Policies").Joins("JOIN account_roles ON account_roles.role_id = roles.id AND account_roles.deleted_at = 0").
			Where("account_roles.account_id = ?", accountID).
			Where(app.Role{OrgID: generics.NewNullString(orgID)}).
			Where("roles.role_type IN ?", []app.RoleType{app.RoleTypeRunner, app.RoleTypeRunnerInstall}).
			Order("roles.id").Find(&existing).Error; err != nil {
			return fmt.Errorf("load runner roles for consolidation: %w", err)
		}
		for _, role := range existing {
			object := orgID
			if role.RoleType == app.RoleTypeRunnerInstall {
				if installID == "" {
					return fmt.Errorf("runner has an install role but does not serve an install")
				}
				if err := validatePrivateRole(tx, accountID, orgID, &role); err != nil {
					return err
				}
				object = permissions.Object(orgID, permissions.KindInstall, installID)
			}
			if len(role.Policies) != 1 {
				return fmt.Errorf("runner role %s requires one policy", role.ID)
			}
			projection, err := role.Policies[0].PermissionProjection()
			if err != nil {
				return err
			}
			expected := pgtype.Hstore{
				permissions.APIPermissionKey: generics.ToPtr(string(permissions.APIRunner)),
				object:                       permissions.PermissionAll.ToStrPtr(),
			}
			if !reflect.DeepEqual(projection, expected) {
				return fmt.Errorf("runner role %s has unexpected grants; refusing to change access", role.ID)
			}
			if role.RoleType == app.RoleTypeRunnerInstall {
				if err := deletePrivateRole(tx, accountID, role.ID); err != nil {
					return err
				}
			} else if err := tx.Unscoped().Where(app.AccountRole{AccountID: accountID, RoleID: role.ID}).Delete(&app.AccountRole{}).Error; err != nil {
				return fmt.Errorf("remove shared runner role assignment: %w", err)
			}
		}
		return nil
	})
}
