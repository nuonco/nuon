package account

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
)

// DeleteServiceAccount removes an account's role bindings, stack roles, tokens, and
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

	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Before deleteAccountRecords: the bindings are how the roles are found.
		if err := authz.DeleteStackInstallRoles(tx, acct.ID); err != nil {
			return err
		}

		return deleteAccountRecords(tx, acct.ID)
	})
}

// Role bindings are hard-deleted: the many2many's OnDelete:CASCADE never fires on a
// soft delete. Soft-deleting tokens and the account is enough to break auth, since
// FindAccount cannot see soft-deleted rows.
func deleteAccountRecords(tx *gorm.DB, accountID string) error {
	if res := tx.Unscoped().
		Where(app.AccountRole{AccountID: accountID}).
		Delete(&app.AccountRole{}); res.Error != nil {
		return errors.Wrap(res.Error, "unable to remove account roles")
	}

	if res := tx.
		Where(app.Token{AccountID: accountID}).
		Delete(&app.Token{}); res.Error != nil {
		return errors.Wrap(res.Error, "unable to delete tokens")
	}

	if res := tx.Delete(&app.Account{ID: accountID}); res.Error != nil {
		return errors.Wrap(res.Error, "unable to delete account")
	}

	return nil
}

// DeleteInstallStackServiceAccounts removes the stack accounts for an install's
// stacks. Must run before the install row is deleted: the account is tied to the
// stack by naming convention, not a foreign key, so nothing cascades to it and the
// IDs are unrecoverable afterwards.
func (c *Client) DeleteInstallStackServiceAccounts(ctx context.Context, installID string) error {
	var stackIDs []string
	if res := c.db.WithContext(ctx).
		Model(&app.InstallStack{}).
		Where(app.InstallStack{InstallID: installID}).
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
