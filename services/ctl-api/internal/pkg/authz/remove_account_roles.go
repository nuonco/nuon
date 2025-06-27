package authz

import (
	"context"

	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (h *Client) RemoveAccountOrgRoles(ctx context.Context, orgID, accountID string) error {
	// Delete all roles for the account in the specified organization
	res := h.db.WithContext(ctx).
		Where(app.AccountRole{
			OrgID:     generics.NewNullString(orgID),
			AccountID: accountID,
		}).
		Delete(&app.AccountRole{})

	if res.Error != nil {
		return errors.Wrap(res.Error, "unable to remove roles for account")
	}

	return nil
}
