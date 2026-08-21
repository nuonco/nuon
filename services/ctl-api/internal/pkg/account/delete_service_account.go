package account

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// DeleteServiceAccount removes a service account and everything that makes it usable
// as a credential: its role bindings, its tokens, and the account row itself.
//
// Idempotent by design. A missing account is success, not an error — callers run this
// from delete workflows that retry, and an entity that never had a service account
// must not block its own teardown.
//
// Role bindings are hard-deleted, matching authz.RemoveAccountOrgRoles. The account
// and its tokens are soft-deleted, which is sufficient to kill authentication: the
// auth middleware resolves a token to an account via FindAccount, and a soft-deleted
// account is invisible to it. The unique index on accounts spans deleted_at, so the
// service-account email is freed for reuse either way.
func (c *Client) DeleteServiceAccount(ctx context.Context, svcAcctID string) error {
	email := ServiceAccountEmail(svcAcctID)

	acct, err := c.FindAccount(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return errors.Wrap(err, "unable to look up service account")
	}

	// Guard against deleting a human account through a path meant for machine
	// identities. FindAccount matches on email, subject, or ID, so a caller passing
	// something unexpected could otherwise reach a real user.
	if acct.AccountType != app.AccountTypeService {
		return errors.Errorf("account %s is not a service account", acct.ID)
	}

	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Hard delete: the many2many association declares OnDelete:CASCADE, but that
		// is a foreign-key constraint and a soft delete never fires it.
		if res := tx.Unscoped().
			Where(app.AccountRole{AccountID: acct.ID}).
			Delete(&app.AccountRole{}); res.Error != nil {
			return errors.Wrap(res.Error, "unable to remove account roles")
		}

		if res := tx.
			Where(app.Token{AccountID: acct.ID}).
			Delete(&app.Token{}); res.Error != nil {
			return errors.Wrap(res.Error, "unable to delete tokens")
		}

		if res := tx.Delete(&app.Account{ID: acct.ID}); res.Error != nil {
			return errors.Wrap(res.Error, "unable to delete account")
		}

		return nil
	})
}
