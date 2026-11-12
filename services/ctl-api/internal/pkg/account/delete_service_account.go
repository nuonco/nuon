package account

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
)

// DeleteServiceAccount removes an account's role bindings, private roles, tokens, and
// the row itself. A missing account is success — delete workflows retry.
func (c *Client) DeleteServiceAccount(ctx context.Context, svcAcctID string) error {
	email := ServiceAccountEmail(svcAcctID)

	acct, err := c.FindAccount(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return errors.Wrap(err, "unable to look up service account")
	}

	// FindAccount matches email, subject, or ID, so an unexpected argument could reach a
	// real user.
	if acct.AccountType != app.AccountTypeService {
		return errors.Errorf("account %s is not a service account", acct.ID)
	}

	return c.DeleteServiceAccountByID(ctx, acct.ID)
}

func (c *Client) DeleteServiceAccountByID(ctx context.Context, accountID string) error {
	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return deleteServiceAccounts(tx, []string{accountID})
	})
}

func (c *Client) RevokeServiceAccountCredentials(ctx context.Context, svcAcctID string) error {
	acct, err := c.FindAccount(ctx, ServiceAccountEmail(svcAcctID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return errors.Wrap(err, "unable to look up service account")
	}
	if acct.AccountType != app.AccountTypeService {
		return errors.Errorf("account %s is not a service account", acct.ID)
	}
	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return revokeAccountCredentials(tx, acct.ID)
	})
}

func deleteServiceAccount(tx *gorm.DB, accountID string) error {
	return deleteServiceAccounts(tx, []string{accountID})
}

func deleteServiceAccounts(tx *gorm.DB, accountIDs []string) error {
	var accts []app.Account
	if err := tx.Unscoped().Select("id", "account_type").
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id IN ?", accountIDs).Order("id").Find(&accts).Error; err != nil {
		return errors.Wrap(err, "unable to lock service accounts for deletion")
	}
	if len(accts) == 0 {
		return nil
	}
	ids := make([]string, 0, len(accts))
	for _, acct := range accts {
		if acct.AccountType != app.AccountTypeService {
			return errors.Errorf("account %s is not a service account", acct.ID)
		}
		ids = append(ids, acct.ID)
	}
	if err := deleteAccountCredentials(tx, ids); err != nil {
		return err
	}
	if err := authz.DeleteManagedServiceAccountRoles(tx, ids); err != nil {
		return err
	}
	if err := authz.DeletePrivateRoles(tx, ids); err != nil {
		return err
	}
	if err := tx.Where("account_id IN ?", ids).Delete(&app.ManagedServiceAccount{}).Error; err != nil {
		return errors.Wrap(err, "unable to remove managed service account bindings")
	}
	return deleteAccountRecords(tx, ids)
}

// Role bindings are hard-deleted: the many2many's OnDelete:CASCADE never fires on a
// soft delete. Soft-deleting tokens and the account is enough to break auth, since
// FindAccount cannot see soft-deleted rows.
func deleteAccountRecords(tx *gorm.DB, accountIDs []string) error {
	if res := tx.Unscoped().
		Where("account_id IN ?", accountIDs).
		Delete(&app.AccountRole{}); res.Error != nil {
		return errors.Wrap(res.Error, "unable to remove account roles")
	}

	if res := tx.Where("id IN ?", accountIDs).Delete(&app.Account{}); res.Error != nil {
		return errors.Wrap(res.Error, "unable to delete accounts")
	}

	return nil
}

// DeleteInstallStackServiceAccounts removes the stack accounts for an install's
// stacks. Must run before the install row is deleted: the account is tied to the
// stack by naming convention, not a foreign key, so nothing cascades to it and the
// IDs are unrecoverable afterwards.
func (c *Client) DeleteInstallStackServiceAccounts(ctx context.Context, installID string) error {
	var stackIDs []string
	if res := c.db.WithContext(ctx).Unscoped().
		Model(&app.InstallStack{}).
		Where(app.InstallStack{InstallID: installID}).
		Order("id").
		Pluck("id", &stackIDs); res.Error != nil {
		return errors.Wrap(res.Error, "unable to list install stacks")
	}
	for _, stackID := range stackIDs {
		if err := c.DeleteServiceAccount(ctx, stackID); err != nil {
			return errors.Wrapf(err, "unable to delete stack service account for stack %s", stackID)
		}
	}
	return nil
}
