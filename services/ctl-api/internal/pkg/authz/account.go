package authz

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// TODO(jm): this entire file should probably live in `pkg/account`
func (m *Client) FetchAccount(ctx context.Context, acctID string) (*app.Account, error) {
	acct := app.Account{}
	res := m.db.WithContext(ctx).
		Preload("Roles").
		Preload("Roles.Org").
		Preload("Roles.Policies").
		First(&acct, "id = ?", acctID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to fetch account %s: %w", acctID, res.Error)
	}

	return &acct, nil
}

func (m *Client) FindAccount(ctx context.Context, emailOrSubjectOrID string) (*app.Account, error) {
	acct := app.Account{}
	res := m.db.WithContext(ctx).
		Preload("Roles").
		Preload("Roles.Org").
		Preload("Roles.Policies").
		Where(app.Account{
			Email: emailOrSubjectOrID,
		}).
		Or(app.Account{
			Subject: emailOrSubjectOrID,
		}).
		Or(app.Account{
			ID: emailOrSubjectOrID,
		}).
		First(&acct)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to find account %s: %w", emailOrSubjectOrID, res.Error)
	}

	return &acct, nil
}

func (m *Client) CreateAccount(ctx context.Context, email, subject string) (*app.Account, error) {
	acct := app.Account{
		Email:       email,
		Subject:     subject,
		AccountType: app.AccountTypeAuth0,
	}

	if err := m.db.WithContext(ctx).
		Create(&acct).Error; err != nil {
		return nil, fmt.Errorf("unable to create account: %w", err)
	}

	return &acct, nil
}

func (m *Client) CreateServiceAccount(ctx context.Context, id string) (*app.Account, error) {
	acct := app.Account{
		Email:       fmt.Sprintf("%s@serviceaccount.nuon.co", id),
		Subject:     id,
		AccountType: app.AccountTypeService,
	}

	if err := m.db.WithContext(ctx).
		Create(&acct).Error; err != nil {
		return nil, fmt.Errorf("unable to create service account: %w", err)
	}

	return &acct, nil
}
