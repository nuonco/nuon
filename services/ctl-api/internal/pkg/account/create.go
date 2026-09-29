package account

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func (m *Client) CreateAccount(ctx context.Context, email, subject string, userJourneys app.UserJourneys) (*app.Account, error) {
	return m.createAccount(ctx, email, subject, app.AccountTypeAuth0, userJourneys)
}

func (m *Client) CreateAuthAccount(ctx context.Context, email, subject string, userJourneys app.UserJourneys) (*app.Account, error) {
	return m.createAccount(ctx, email, subject, app.AccountTypeAuth, userJourneys)
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

func NoUserJourneys() app.UserJourneys {
	return app.UserJourneys{}
}

func DefaultEvaluationJourneyWithAttribution(attribution map[string]interface{}, completionSource string) app.UserJourneys {
	journey := DefaultEvaluationJourney(completionSource)

	if len(attribution) > 0 && len(journey) > 0 && len(journey[0].Steps) > 0 {
		for i, step := range journey[0].Steps {
			if step.Name == "account_created" {
				if journey[0].Steps[i].Metadata == nil {
					journey[0].Steps[i].Metadata = make(map[string]interface{})
				}
				journey[0].Steps[i].Metadata["attribution"] = attribution
				break
			}
		}
	}

	return journey
}
