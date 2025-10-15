package helpers

import (
	"context"
	"fmt"
	"time"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// loadAccountWithJourneys loads account with appropriate preloads based on need
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

// StepUpdate defines what can be updated in a journey step
type StepUpdate struct {
	Complete         bool
	CompletedAt      *time.Time
	CompletionMethod string
	CompletionSource string
	Metadata         map[string]interface{}
}

// updateJourneyStepIfIncomplete finds and updates a journey step only if it's incomplete
func (h *Helpers) updateJourneyStepIfIncomplete(account *app.Account, journeyName, stepName string, update StepUpdate) bool {
	for i, journey := range account.UserJourneys {
		if journey.Name == journeyName {
			for j, step := range journey.Steps {
				if step.Name == stepName && !step.Complete {
					// Update completion status
					account.UserJourneys[i].Steps[j].Complete = update.Complete

					// Update completion tracking fields
					if update.Complete {
						account.UserJourneys[i].Steps[j].CompletedAt = update.CompletedAt
						account.UserJourneys[i].Steps[j].CompletionMethod = update.CompletionMethod
						account.UserJourneys[i].Steps[j].CompletionSource = update.CompletionSource
					}

					// Update metadata (merge with existing)
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

// saveAccountJourneys saves only the user_journeys field to database
func (h *Helpers) saveAccountJourneys(ctx context.Context, account *app.Account) error {
	if err := h.db.WithContext(ctx).Select("user_journeys").Save(account).Error; err != nil {
		return fmt.Errorf("unable to update user journey: %w", err)
	}
	return nil
}

// UpdateJourneyStepParams provides flexible parameters for journey step updates
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

// updateUserJourneyStepIfIncomplete is the consolidated method that handles all journey step updates
func (h *Helpers) updateUserJourneyStepIfIncomplete(ctx context.Context, params UpdateJourneyStepParams) error {
	// 1. Load account (with roles only if needed)
	account, err := h.loadAccountWithJourneys(ctx, params.AccountID, params.NeedsRoleData)
	if err != nil {
		return err
	}

	// 2. First-org specific validation
	if params.StepName == "org_created" && len(account.OrgIDs) > 1 {
		return nil // Not first org, skip update
	}

	// 3. Prepare update with completion tracking
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
		return nil // No changes needed
	}

	// 4. Save changes
	return h.saveAccountJourneys(ctx, account)
}

// buildNavigationMetadata creates metadata for navigation purposes
func buildNavigationMetadata(appID, installID, orgID *string) map[string]interface{} {
	metadata := make(map[string]interface{}) // Never return nil

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

// UpdateUserJourneyStepForFirstOrg updates the org_created step when user creates their first org
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

// UpdateUserJourneyStepForCLIInstalled updates the cli_installed step when user installs the CLI
func (h *Helpers) UpdateUserJourneyStepForCLIInstalled(ctx context.Context, accountID string) error {
	return h.updateUserJourneyStepIfIncomplete(ctx, UpdateJourneyStepParams{
		AccountID:        accountID,
		JourneyName:      "evaluation",
		StepName:         "cli_installed",
		Complete:         true,
		CompletionMethod: "auto",
		CompletionSource: "cli", // Detected when CLI makes first API call
		Metadata:         buildNavigationMetadata(nil, nil, nil),
		NeedsRoleData:    false,
	})
}

// UpdateUserJourneyStepForFirstAppCreate updates the app_created step when user creates their first app
func (h *Helpers) UpdateUserJourneyStepForFirstAppCreate(ctx context.Context, accountID, appID string) error {
	return h.updateUserJourneyStepIfIncomplete(ctx, UpdateJourneyStepParams{
		AccountID:        accountID,
		JourneyName:      "evaluation",
		StepName:         "app_created",
		Complete:         true,
		CompletionMethod: "auto",
		CompletionSource: "api", // Triggered by app creation API
		Metadata:         buildNavigationMetadata(&appID, nil, nil),
		NeedsRoleData:    false,
	})
}

// UpdateUserJourneyStepForFirstAppSync updates the app_synced step when user syncs their first app config
func (h *Helpers) UpdateUserJourneyStepForFirstAppSync(ctx context.Context, accountID, appID string) error {
	return h.updateUserJourneyStepIfIncomplete(ctx, UpdateJourneyStepParams{
		AccountID:        accountID,
		JourneyName:      "evaluation",
		StepName:         "app_synced",
		Complete:         true,
		CompletionMethod: "auto",
		CompletionSource: "cli", // Triggered by `nuon apps sync`
		Metadata:         buildNavigationMetadata(&appID, nil, nil),
		NeedsRoleData:    false,
	})
}

// UpdateUserJourneyStepForFirstInstallCreate updates the install_created step when user creates their first install
func (h *Helpers) UpdateUserJourneyStepForFirstInstallCreate(ctx context.Context, accountID, installID string) error {
	return h.updateUserJourneyStepIfIncomplete(ctx, UpdateJourneyStepParams{
		AccountID:        accountID,
		JourneyName:      "evaluation",
		StepName:         "install_created",
		Complete:         true,
		CompletionMethod: "auto",
		CompletionSource: "dashboard", // Usually created via dashboard
		Metadata:         buildNavigationMetadata(nil, &installID, nil),
		NeedsRoleData:    false,
	})
}

// UpdateUserJourneyStep provides a general method for updating any user journey step
func (h *Helpers) UpdateUserJourneyStep(ctx context.Context, accountID, journeyName, stepName string, complete bool) error {
	return h.updateUserJourneyStepIfIncomplete(ctx, UpdateJourneyStepParams{
		AccountID:        accountID,
		JourneyName:      journeyName,
		StepName:         stepName,
		Complete:         complete,
		CompletionMethod: "manual", // Generic method assumes manual completion
		CompletionSource: "api",    // Coming through API
		Metadata:         make(map[string]interface{}),
		NeedsRoleData:    false,
	})
}
