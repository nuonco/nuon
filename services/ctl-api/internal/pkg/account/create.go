package account

import (
	"context"
	"fmt"

	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func (m *Client) CreateAuthAccount(ctx context.Context, email, subject string, userJourneys app.UserJourneys) (*app.Account, error) {
	return m.createAccount(ctx, email, subject, app.AccountTypeAuth, userJourneys)
}

// CreateTestAccount creates an integration or canary account for subject. Canary
// subjects are reused across runs, so an existing row is taken over in place.
// Test accounts are not identified to analytics.
func (m *Client) CreateTestAccount(ctx context.Context, subject string, accountType app.AccountType) (*app.Account, error) {
	if accountType != app.AccountTypeIntegration && accountType != app.AccountTypeCanary {
		return nil, fmt.Errorf("unsupported test account type %q", accountType)
	}
	acct := app.Account{
		Email:       fmt.Sprintf("%s@nuon.co", subject),
		Subject:     subject,
		AccountType: accountType,
	}
	if err := m.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "email"}, {Name: "subject"}, {Name: "deleted_at"}},
			UpdateAll: true,
		}).
		Create(&acct).Error; err != nil {
		return nil, fmt.Errorf("unable to create %s account: %w", accountType, err)
	}
	return &acct, nil
}

func (m *Client) createAccount(ctx context.Context, email, subject string, accountType app.AccountType, userJourneys app.UserJourneys) (*app.Account, error) {
	acct := app.Account{
		Email:        email,
		Subject:      subject,
		AccountType:  accountType,
		UserJourneys: userJourneys,
	}

	if err := m.db.WithContext(ctx).
		Create(&acct).Error; err != nil {
		return nil, fmt.Errorf("unable to create account: %w", err)
	}

	ctx = cctx.SetAccountContext(ctx, &acct)
	m.analyticsClient.Identify(ctx)
	return &acct, nil
}

// DefaultEvaluationJourney returns the evaluation journey for self-signup users
// This is the 6-step journey: account_created, org_created, cli_installed, app_created, app_synced, install_created
func DefaultEvaluationJourney(completionSource string) app.UserJourneys {
	return app.UserJourneys{
		{
			Name:  "evaluation",
			Title: "Getting Started",
			Steps: []app.UserJourneyStep{
				{
					Name:             "account_created",
					Title:            "Create an account",
					Complete:         false,
					CompletedAt:      nil,
					CompletionMethod: "",
					CompletionSource: "",
					Metadata:         make(map[string]interface{}),
				},
				{
					Name:             "org_created",
					Title:            "Create an organization",
					Complete:         false,
					CompletedAt:      nil,
					CompletionMethod: "",
					CompletionSource: "",
					Metadata:         make(map[string]interface{}),
				},
				{
					Name:             "cli_installed",
					Title:            "Install the CLI",
					Complete:         false,
					CompletedAt:      nil,
					CompletionMethod: "",
					CompletionSource: "",
					Metadata:         make(map[string]interface{}),
				},
				{
					Name:             "app_created",
					Title:            "Create an app",
					Complete:         false,
					CompletedAt:      nil,
					CompletionMethod: "",
					CompletionSource: "",
					Metadata:         make(map[string]interface{}),
				},
				{
					Name:             "app_synced",
					Title:            "Sync app configuration",
					Complete:         false,
					CompletedAt:      nil,
					CompletionMethod: "",
					CompletionSource: "",
					Metadata:         make(map[string]interface{}),
				},
				{
					Name:             "install_created",
					Title:            "Create an install",
					Complete:         false,
					CompletedAt:      nil,
					CompletionMethod: "",
					CompletionSource: "",
					Metadata:         make(map[string]interface{}),
				},
			},
		},
	}
}

// NoUserJourneys returns an empty journey slice for invited/support users
func NoUserJourneys() app.UserJourneys {
	return app.UserJourneys{}
}
