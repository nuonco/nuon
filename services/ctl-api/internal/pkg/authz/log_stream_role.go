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

func logStreamWriteRole(orgID, logStreamID string) *app.Role {
	return &app.Role{
		OrgID:       generics.NewNullString(orgID),
		RoleType:    app.RoleTypeLogStream,
		Title:       "Log stream",
		Description: "Write logs to a single log stream.",
		Policies: []app.Policy{{
			OrgID: generics.NewNullString(orgID),
			Name:  app.PolicyNameLogStream,
			Permissions: pgtype.Hstore{
				permissions.Object(orgID, permissions.KindLogStream, logStreamID): permissions.PermissionCreate.ToStrPtr(),
			},
		}},
	}
}

func (h *Client) EnsureLogStreamWriteRole(ctx context.Context, orgID, logStreamID, accountID string) error {
	ctx = cctx.SetAccountIDContext(ctx, accountID)
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var acct app.Account
		if err := tx.Select("id", "account_type", "subject").Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(app.Account{ID: accountID}).Take(&acct).Error; err != nil {
			return fmt.Errorf("lock log stream service account: %w", err)
		}
		if acct.AccountType != app.AccountTypeService || acct.Subject != logStreamID {
			return fmt.Errorf("account does not match the log stream service identity")
		}

		want := logStreamWriteRole(orgID, logStreamID)
		var existing struct {
			RoleID string
		}
		err := tx.Model(&app.AccountRole{}).
			Select("account_roles.role_id").
			Joins("JOIN roles ON roles.id = account_roles.role_id AND roles.deleted_at = 0").
			Where(app.AccountRole{AccountID: accountID}).
			Where("roles.role_type = ?", app.RoleTypeLogStream).
			Take(&existing).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("load log stream role assignment: %w", err)
		}
		want.ID = existing.RoleID
		if err := ensurePrivateRole(tx, accountID, want); err != nil {
			return err
		}
		if err := tx.Where(app.AccountRole{AccountID: accountID}).
			Where("role_id <> ?", want.ID).
			Delete(&app.AccountRole{}).Error; err != nil {
			return fmt.Errorf("remove broad log stream account roles: %w", err)
		}
		return nil
	})
}
