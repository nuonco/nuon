package account

import (
	"context"
	"fmt"
	"time"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

func (m *Client) CreateAccount(ctx context.Context, email, subject string, userJourneys app.UserJourneys) (*app.Account, error) {
	acct := app.Account{
		Email:        email,
		Subject:      subject,
		AccountType:  app.AccountTypeAuth0,
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

// DefaultEvaluationJourneyWithAutoOrg returns the evaluation journey for self-signup users with auto-org creation
func DefaultEvaluationJourneyWithAutoOrg() app.UserJourneys {
	now := time.Now().UTC()

	return app.UserJourneys{
		{
			Name:  "evaluation",
			Title: "Getting Started",
			Steps: []app.UserJourneyStep{
				{
					Name:             "account_created",
					Title:            "Create an account",
					Complete:         true,
					CompletedAt:      &now,
					CompletionMethod: "auto",
					CompletionSource: "system",
					Metadata:         make(map[string]interface{}),
				},
				{
					Name:             "org_created",
					Title:            "Create an organization",
					Complete:         true,
					CompletedAt:      &now,
					CompletionMethod: "auto",
					CompletionSource: "system",
					Metadata:         make(map[string]interface{}),
				},
				{
					Name:             "cli_installed",
					Title:            "Install the Nuon CLI",
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
					Title:            "Sync the app config",
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

// DefaultEvaluationJourney returns the evaluation journey for self-signup users without auto-org creation
func DefaultEvaluationJourney() app.UserJourneys {
	now := time.Now().UTC()

	return app.UserJourneys{
		{
			Name:  "evaluation",
			Title: "Getting Started",
			Steps: []app.UserJourneyStep{
				{
					Name:             "account_created",
					Title:            "Create an account",
					Complete:         true,
					CompletedAt:      &now,
					CompletionMethod: "auto",
					CompletionSource: "system",
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
					Title:            "Install the Nuon CLI",
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
					Title:            "Sync the app config",
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
