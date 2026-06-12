package activities

import (
	"context"
	"fmt"
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/salesforce"
)

type SendTrialSignupRequest struct {
	OrgID  string `json:"org_id" validate:"required"`
	Source string `json:"source"`
}

// @temporal-gen-v2 activity
func (a *Activities) SendTrialSignup(ctx context.Context, req SendTrialSignupRequest) (bool, error) {
	if !a.salesforce.Enabled() {
		return false, nil
	}

	org, err := a.getOrg(ctx, req.OrgID)
	if err != nil {
		return false, fmt.Errorf("unable to get org: %w", err)
	}

	var account app.Account
	res := a.db.WithContext(ctx).
		Preload("Roles").
		Preload("Roles.Org").
		Preload("Roles.Policies").
		Preload("Identities").
		First(&account, "id = ?", org.CreatedByID)
	if res.Error != nil {
		return false, fmt.Errorf("unable to get org creator account: %w", res.Error)
	}

	switch account.AccountType {
	case app.AccountTypeService, app.AccountTypeCanary, app.AccountTypeIntegration:
		return false, nil
	}

	if len(account.OrgIDs) != 1 {
		return false, nil
	}

	signup := buildTrialSignup(&account, org, req.Source)
	if err := a.salesforce.SendTrialSignup(ctx, signup); err != nil {
		return false, fmt.Errorf("unable to send trial signup: %w", err)
	}

	return true, nil
}

func buildTrialSignup(account *app.Account, org *app.Org, source string) salesforce.TrialSignup {
	var fullName string
	for _, identity := range account.Identities {
		if identity.Name != "" {
			fullName = identity.Name
			break
		}
	}

	nameParts := strings.Fields(fullName)
	var firstName, lastName string
	if len(nameParts) > 0 {
		firstName = nameParts[0]
		lastName = strings.Join(nameParts[1:], " ")
	}
	if lastName == "" {
		lastName = "ULN"
	}

	notes := fmt.Sprintf("Org: %s", org.Name)
	if source != "" {
		notes = fmt.Sprintf("Created via %s. %s", source, notes)
	}

	return salesforce.TrialSignup{
		FirstName: firstName,
		LastName:  lastName,
		Email:     account.Email,
		Notes:     notes,
		Subject:   salesforce.TrialSignupSubject,
	}
}
