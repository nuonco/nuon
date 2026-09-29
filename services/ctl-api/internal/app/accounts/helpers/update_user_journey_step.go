package helpers

import (
	"context"
	"fmt"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (h *Helpers) loadAccountWithJourneys(ctx context.Context, accountID string, needsRoles bool) (*app.Account, error) {
	var account app.Account
	query := h.db.WithContext(ctx)

	if needsRoles {
		query = query.Preload("Roles").Preload("Roles.Org").Preload("Roles.Policies")
	}

	if err := query.Where("id = ?", accountID).First(&account).Error; err != nil {
		return nil, fmt.Errorf("unable to get account: %w", err)
	}

	return &account, nil
}

type StepUpdate struct {
	Complete         bool
	CompletedAt      *time.Time
	CompletionMethod string
	CompletionSource string
	Metadata         map[string]interface{}
}

func (h *Helpers) updateJourneyStepIfIncomplete(account *app.Account, journeyName, stepName string, update StepUpdate) bool {
	for i, journey := range account.UserJourneys {
		if journey.Name == journeyName {
			for j, step := range journey.Steps {
				if step.Name == stepName && !step.Complete {
					account.UserJourneys[i].Steps[j].Complete = update.Complete

					if update.Complete {
						account.UserJourneys[i].Steps[j].CompletedAt = update.CompletedAt
						account.UserJourneys[i].Steps[j].CompletionMethod = update.CompletionMethod
						account.UserJourneys[i].Steps[j].CompletionSource = update.CompletionSource
					}

					if update.Metadata != nil {
						if account.UserJourneys[i].Steps[j].Metadata == nil {
							account.UserJourneys[i].Steps[j].Metadata = make(map[string]interface{})
						}
						for k, v := range update.Metadata {
							account.UserJourneys[i].Steps[j].Metadata[k] = v
						}
					}

					return true
				}
			}
			break
		}
	}
	return false
}

func (h *Helpers) saveAccountJourneys(ctx context.Context, account *app.Account) error {
	if err := h.db.WithContext(ctx).Select("user_journeys").Save(account).Error; err != nil {
		return fmt.Errorf("unable to update user journey: %w", err)
	}
	return nil
}

type UpdateJourneyStepParams struct {
	AccountID        string
	JourneyName      string
	StepName         string
	Complete         bool
	CompletionMethod string
	CompletionSource string
	Metadata         map[string]interface{}
	NeedsRoleData    bool
}

func (h *Helpers) updateUserJourneyStepIfIncomplete(ctx context.Context, params UpdateJourneyStepParams) error {
	account, err := h.loadAccountWithJourneys(ctx, params.AccountID, params.NeedsRoleData)
	if err != nil {
		return err
	}

	if params.StepName == "org_created" && len(account.OrgIDs) > 1 {
		return nil
	}

	now := time.Now().UTC()
	update := StepUpdate{
		Complete:         params.Complete,
		CompletedAt:      &now,
		CompletionMethod: params.CompletionMethod,
		CompletionSource: params.CompletionSource,
		Metadata:         params.Metadata,
	}

	updated := h.updateJourneyStepIfIncomplete(account, params.JourneyName, params.StepName, update)
	if !updated {
		return nil
	}

	return h.saveAccountJourneys(ctx, account)
}

func buildNavigationMetadata(appID, installID, orgID *string) map[string]interface{} {
	metadata := make(map[string]interface{})

	if appID != nil && *appID != "" {
		metadata["app_id"] = *appID
	}
	if installID != nil && *installID != "" {
		metadata["install_id"] = *installID
	}
	if orgID != nil && *orgID != "" {
		metadata["org_id"] = *orgID
	}

	return metadata
}

func (h *Helpers) UpdateUserJourneyStepForFirstOrg(ctx context.Context, accountID, orgID string) error {
	return h.updateUserJourneyStepIfIncomplete(ctx, UpdateJourneyStepParams{
		AccountID:        accountID,
		JourneyName:      "evaluation",
		StepName:         "org_created",
		Complete:         true,
		CompletionMethod: "auto",
		CompletionSource: "system",
		Metadata:         buildNavigationMetadata(nil, nil, &orgID),
		NeedsRoleData:    true,
	})
}

func (h *Helpers) UpdateUserJourneyStepForFirstAppCreate(ctx context.Context, accountID, appID string) error {
	return h.updateUserJourneyStepIfIncomplete(ctx, UpdateJourneyStepParams{
		AccountID:        accountID,
		JourneyName:      "evaluation",
		StepName:         "app_created",
		Complete:         true,
		CompletionMethod: "auto",
		CompletionSource: "api",
		Metadata:         buildNavigationMetadata(&appID, nil, nil),
		NeedsRoleData:    false,
	})
}

func (h *Helpers) UpdateUserJourneyStepForFirstInstallCreate(ctx context.Context, accountID, installID string) error {
	return h.updateUserJourneyStepIfIncomplete(ctx, UpdateJourneyStepParams{
		AccountID:        accountID,
		JourneyName:      "evaluation",
		StepName:         "install_created",
		Complete:         true,
		CompletionMethod: "auto",
		CompletionSource: "dashboard",
		Metadata:         buildNavigationMetadata(nil, &installID, nil),
		NeedsRoleData:    false,
	})
}

func (h *Helpers) UpdateUserJourneyStep(ctx context.Context, accountID, journeyName, stepName string, complete bool) error {
	return h.updateUserJourneyStepIfIncomplete(ctx, UpdateJourneyStepParams{
		AccountID:        accountID,
		JourneyName:      journeyName,
		StepName:         stepName,
		Complete:         complete,
		CompletionMethod: "manual",
		CompletionSource: "api",
		Metadata:         make(map[string]interface{}),
		NeedsRoleData:    false,
	})
}

func (h *Helpers) UpdateUserJourneyStepForCLIInstalled(ctx context.Context, accountID string) error {
	return h.updateUserJourneyStepIfIncomplete(ctx, UpdateJourneyStepParams{
		AccountID:        accountID,
		JourneyName:      "evaluation",
		StepName:         "cli_installed",
		Complete:         true,
		CompletionMethod: "auto",
		CompletionSource: "cli",
		Metadata:         make(map[string]interface{}),
		NeedsRoleData:    false,
	})
}

func (h *Helpers) UpdateUserJourneyStepForFirstAppSync(ctx context.Context, accountID, appID string) error {
	return h.updateUserJourneyStepIfIncomplete(ctx, UpdateJourneyStepParams{
		AccountID:        accountID,
		JourneyName:      "evaluation",
		StepName:         "app_synced",
		Complete:         true,
		CompletionMethod: "auto",
		CompletionSource: "cli",
		Metadata:         buildNavigationMetadata(&appID, nil, nil),
		NeedsRoleData:    false,
	})
}
