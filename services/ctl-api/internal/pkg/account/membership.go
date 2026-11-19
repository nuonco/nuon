package account

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
)

func (c *Client) CreateOrgServiceAccount(ctx context.Context, orgID, subject, name string, role app.RoleType) (*app.Account, error) {
	var acct *app.Account
	err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockOrg(tx, orgID); err != nil {
			return err
		}
		client := c.withDB(tx)
		var err error
		acct, err = client.EnsureServiceAccount(ctx, subject, name)
		if err != nil {
			return err
		}
		if err := authz.RequireUserManaged(tx, acct.ID); err != nil {
			return err
		}
		return client.authzClient.AddAccountOrgRole(ctx, role, orgID, acct.ID)
	})
	return acct, err
}

func lockOrg(tx *gorm.DB, orgID string) error {
	if orgID == "" {
		return fmt.Errorf("org ID is required")
	}
	// KEY SHARE conflicts only with org cleanup's FOR UPDATE, not with org updates.
	return tx.Select("id").Clauses(clause.Locking{Strength: "KEY SHARE"}).Where(app.Org{ID: orgID}).Take(&app.Org{}).Error
}

func (c *Client) AddOrgRole(ctx context.Context, orgID, accountID string, role app.RoleType) error {
	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockOrg(tx, orgID); err != nil {
			return err
		}
		if err := authz.RequireUserManaged(tx, accountID); err != nil {
			return err
		}
		return c.withDB(tx).authzClient.AddAccountOrgRole(ctx, role, orgID, accountID)
	})
}

func (c *Client) SetOrgRole(ctx context.Context, orgID, accountID string, role app.RoleType) error {
	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockOrg(tx, orgID); err != nil {
			return err
		}
		if err := authz.RequireUserManaged(tx, accountID); err != nil {
			return err
		}
		return c.withDB(tx).authzClient.SetAccountOrgRole(ctx, orgID, accountID, role)
	})
}

func (c *Client) RenameServiceAccount(ctx context.Context, accountID, name string) error {
	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := authz.RequireUserManaged(tx, accountID); err != nil {
			return err
		}
		return tx.Model(&app.Account{}).Where(app.Account{ID: accountID, AccountType: app.AccountTypeService}).Update("name", name).Error
	})
}

func (c *Client) RemoveOrgAccess(ctx context.Context, orgID, accountID string, revokeAll bool) error {
	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockOrg(tx, orgID); err != nil {
			return err
		}
		if err := authz.RequireUserManaged(tx, accountID); err != nil {
			return err
		}
		client := c.withDB(tx)
		if err := client.authzClient.RemoveAccountOrgRoles(ctx, orgID, accountID); err != nil {
			return err
		}
		if revokeAll {
			return revokeAccountCredentials(tx, accountID)
		}
		return client.RevokeOrgTokens(ctx, orgID, accountID)
	})
}
