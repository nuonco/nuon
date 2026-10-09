package account

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
)

// DeleteServiceAccount removes an account's role bindings, stack roles, tokens, and
// the row itself. A missing account is success — delete workflows retry.
func (c *Client) DeleteServiceAccount(ctx context.Context, svcAcctID string) error {
	acct, err := c.FindAccount(ctx, svcAcctID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		acct, err = c.FindAccount(ctx, ServiceAccountEmail(svcAcctID))
	}
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
		if err := revokeAccountCredentials(tx, acct.ID); err != nil {
			return err
		}

		// Before deleteAccountRecords: the bindings are how the roles are found.
		if err := authz.DeleteStackInstallRoles(tx, acct.ID); err != nil {
			return err
		}

		return deleteAccountRecords(tx, acct)
	})
}

func (c *Client) RemoveServiceAccountFromOrg(ctx context.Context, orgID, accountID string) error {
	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := revokeAccountCredentials(tx, accountID); err != nil {
			return err
		}
		if err := authz.RemoveAccountOrgRoles(tx, orgID, accountID); err != nil {
			return err
		}
		if res := tx.Where(app.OAuthClient{OrgID: generics.NewNullString(orgID), AccountID: generics.NewNullString(accountID)}).
			Delete(&app.OAuthClient{}); res.Error != nil {
			return errors.Wrap(res.Error, "unable to delete org service account oauth clients")
		}
		return nil
	})
}

// Role bindings are hard-deleted: the many2many's OnDelete:CASCADE never fires on a
// soft delete. Soft-deleting tokens and the account is enough to break auth, since
// FindAccount cannot see soft-deleted rows.
func deleteAccountRecords(tx *gorm.DB, acct *app.Account) error {
	if res := tx.Unscoped().
		Where(app.AccountRole{AccountID: acct.ID}).
		Delete(&app.AccountRole{}); res.Error != nil {
		return errors.Wrap(res.Error, "unable to remove account roles")
	}

	if res := tx.Unscoped().Where(app.OrgInvite{Email: acct.Email}).
		Delete(&app.OrgInvite{}); res.Error != nil {
		return errors.Wrap(res.Error, "unable to remove account invites")
	}

	if res := tx.Where(app.OAuthClient{AccountID: generics.NewNullString(acct.ID)}).
		Delete(&app.OAuthClient{}); res.Error != nil {
		return errors.Wrap(res.Error, "unable to delete account oauth clients")
	}

	if res := tx.Delete(&app.Account{ID: acct.ID}); res.Error != nil {
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
