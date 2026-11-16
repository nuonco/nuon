package authz

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func logStreamRole(orgID, logStreamID string) *app.Role {
	return &app.Role{
		OrgID:       generics.NewNullString(orgID),
		RoleType:    app.RoleTypeLogStream,
		Version:     app.RoleVersionV2,
		Title:       "Log stream",
		Description: "Append to one log stream.",
		Policies: []app.Policy{{
			OrgID: generics.NewNullString(orgID),
			Name:  app.PolicyNameLogStream,
			Permissions: pgtype.Hstore{
				permissions.APIPermissionKey:                                      generics.ToPtr(string(permissions.APIRunner)),
				permissions.Object(orgID, permissions.KindLogStream, logStreamID): permissions.PermissionCreate.ToStrPtr(),
			},
		}},
	}
}

func (h *Client) EnsureLogStreamRole(ctx context.Context, orgID, logStreamID, accountID string) error {
	if orgID == "" || logStreamID == "" || accountID == "" {
		return fmt.Errorf("log stream role requires an org, log stream, and account")
	}
	ctx = cctx.SetAccountIDContext(ctx, accountID)
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var acct struct{ ID string }
		if err := tx.Model(&app.Account{}).Select("id").Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(app.Account{ID: accountID}).Take(&acct).Error; err != nil {
			return fmt.Errorf("lock log stream account: %w", err)
		}
		var existing app.Role
		err := tx.Joins("JOIN account_roles ON account_roles.role_id = roles.id AND account_roles.deleted_at = 0").
			Where("account_roles.account_id = ?", accountID).
			Where(app.Role{OrgID: generics.NewNullString(orgID), RoleType: app.RoleTypeLogStream}).
			First(&existing).Error
		if err == nil {
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("load log stream role: %w", err)
		}
		return ensurePrivateRole(tx, accountID, logStreamRole(orgID, logStreamID))
	})
}
