package account

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
)

// DeleteServiceAccount removes a service account and everything that makes it usable
// as a credential: role bindings, its per-install stack roles, tokens, and the
// account row.
//
// A missing account is success: callers run this from delete workflows that retry,
// and an entity that never had a service account must not block its own teardown.
// The unique index on accounts spans deleted_at, so the email is freed for reuse.
func (c *Client) DeleteServiceAccount(ctx context.Context, svcAcctID string) error {
	email := ServiceAccountEmail(svcAcctID)

	acct, err := c.FindAccount(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return errors.Wrap(err, "unable to look up service account")
	}

	// FindAccount matches on email, subject, or ID, so a caller passing something
	// unexpected could otherwise reach a real user.
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

// deleteAccountRecords removes everything that makes an account usable as a credential.
//
// Role bindings are hard-deleted: the many2many declares OnDelete:CASCADE, a
// foreign-key constraint a soft delete never fires. Soft-deleting the tokens and the
// account is enough to break auth, since the middleware resolves tokens through
// FindAccount, which cannot see soft-deleted rows.
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
