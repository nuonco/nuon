package authz

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (h *Client) RemoveAccountOrgRoles(ctx context.Context, orgID, accountID string) error {
	// Struct filters drop zero values, so an empty id would delete every role in the org.
	if accountID == "" || orgID == "" {
		return gorm.ErrRecordNotFound
	}

	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if _, err := lockRoleSubject(tx, accountID); err != nil {
			return err
		}
		// Hard delete all roles for the account in the specified organization
		res := tx.
			Unscoped().
			Where("org_id = ? AND account_id = ?", orgID, accountID).
			Delete(&app.AccountRole{})

		if res.Error != nil {
			return errors.Wrap(res.Error, "unable to remove roles for account")
		}

		// This allows re-inviting the same email address after removal
		var account app.Account
		if err := tx.Select("email").Where("id = ?", accountID).First(&account).Error; err != nil {
			return errors.Wrap(err, "unable to find account for invite cleanup")
		}

		// Hard delete the invite records using Unscoped()
		inviteRes := tx.
			Unscoped().
			Where(&app.OrgInvite{
				OrgID: orgID,
				Email: account.Email,
			}).
			Delete(&app.OrgInvite{})

		if inviteRes.Error != nil {
			return errors.Wrap(inviteRes.Error, "unable to remove invites for account")
		}

		return nil
	})
}
