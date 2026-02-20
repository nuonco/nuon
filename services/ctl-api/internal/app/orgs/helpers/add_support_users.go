package helpers

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type SupportUserResult struct {
	Email         string
	Success       bool
	AlreadyExists bool
	Error         error
}

func (h *Helpers) AddSupportUsersToOrg(ctx context.Context, org *app.Org) ([]SupportUserResult, error) {
	ctx = cctx.SetAccountIDContext(ctx, org.CreatedByID)

	results := make([]SupportUserResult, 0, len(defaultSupportUsers))

	for _, user := range defaultSupportUsers {
		subject := user[0]
		email := user[1]

		result := h.addSingleSupportUser(ctx, subject, email, org.ID)
		results = append(results, result)
	}

	return results, nil
}

func (h *Helpers) addSingleSupportUser(ctx context.Context, subject, email, orgID string) SupportUserResult {
	result := SupportUserResult{
		Email: email,
	}

	acct, err := h.acctClient.FindAccount(ctx, email)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			result.Error = err
			return result
		}

		acct, err = h.acctClient.CreateAccount(ctx, email, subject, account.NoUserJourneys())
		if err != nil {
			result.Error = err
			return result
		}
	}

	err = h.authzClient.AddAccountOrgRole(ctx, app.RoleTypeOrgAdmin, orgID, acct.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) || (err != nil && err.Error() == "role already assigned") {
			result.AlreadyExists = true
			result.Success = true
			return result
		}
		result.Error = err
		return result
	}

	result.Success = true
	return result
}

func (h *Helpers) RemoveSupportUsersFromOrg(ctx context.Context, org *app.Org) error {
	ctx = cctx.SetAccountIDContext(ctx, org.CreatedByID)

	for _, user := range defaultSupportUsers {
		email := user[1]

		acct, err := h.acctClient.FindAccount(ctx, email)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			continue
		}

		if err := h.authzClient.RemoveAccountOrgRoles(ctx, org.ID, acct.ID); err != nil {
			return err
		}
	}

	return nil
}
