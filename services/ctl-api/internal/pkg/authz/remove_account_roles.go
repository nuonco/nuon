package authz

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (h *Client) RemoveAccountOrgRoles(ctx context.Context, orgID, accountID string) error {
	return RemoveAccountOrgRoles(h.db.WithContext(ctx), orgID, accountID)
}

func RemoveAccountOrgRoles(db *gorm.DB, orgID, accountID string) error {
	// Hard delete all roles for the account in the specified organization
	res := db.
		Unscoped().
		Where(app.AccountRole{
			OrgID:     generics.NewNullString(orgID),
			AccountID: accountID,
		}).
		Delete(&app.AccountRole{})

	if res.Error != nil {
		return errors.Wrap(res.Error, "unable to remove roles for account")
	}

	// This allows re-inviting the same email address after removal
	var account app.Account
	if err := db.Select("email").Where("id = ?", accountID).First(&account).Error; err != nil {
		return errors.Wrap(err, "unable to find account for invite cleanup")
	}

	// Hard delete the invite records using Unscoped()
	inviteRes := db.
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
}
