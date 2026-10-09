package activities

import (
	"context"
	"fmt"

	posthog "github.com/posthog/posthog-go"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const orgCreatedEvent = "org_created"

type SendOrgCreatedEventRequest struct {
	OrgID  string `json:"org_id" validate:"required"`
	Source string `json:"source"`
}

// @temporal-gen-v2 activity
func (a *Activities) SendOrgCreatedEvent(ctx context.Context, req SendOrgCreatedEventRequest) (bool, error) {
	ph := a.productAnalytics.PostHog()
	if ph == nil {
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
		Preload("Identities").
		First(&account, "id = ?", org.CreatedByID)
	if res.Error != nil {
		return false, fmt.Errorf("unable to get org creator account: %w", res.Error)
	}

	switch account.AccountType {
	case app.AccountTypeService, app.AccountTypeCanary, app.AccountTypeIntegration:
		return false, nil
	}

	groupIdentify, capture := buildOrgCreatedEvents(&account, org, req.Source)
	if err := ph.Enqueue(groupIdentify); err != nil {
		return false, fmt.Errorf("unable to enqueue org group identify: %w", err)
	}
	if err := ph.Enqueue(capture); err != nil {
		return false, fmt.Errorf("unable to enqueue org created event: %w", err)
	}

	return true, nil
}

func buildOrgCreatedEvents(account *app.Account, org *app.Org, source string) (posthog.GroupIdentify, posthog.Capture) {
	distinctID := account.Email
	if distinctID == "" {
		distinctID = account.ID
	}

	var name string
	for _, identity := range account.Identities {
		if identity.Name != "" {
			name = identity.Name
			break
		}
	}

	personSet := map[string]any{
		"email":      account.Email,
		"account_id": account.ID,
	}
	if name != "" {
		personSet["name"] = name
	}

	tags := []string(org.Tags)

	groupIdentify := posthog.GroupIdentify{
		Type:      "organization",
		Key:       org.ID,
		Timestamp: org.CreatedAt,
		Properties: posthog.NewProperties().
			Set("name", org.Name).
			Set("org_type", string(org.OrgType)).
			Set("tags", tags).
			Set("created_at", org.CreatedAt),
	}

	capture := posthog.Capture{
		DistinctId: distinctID,
		Event:      orgCreatedEvent,
		Timestamp:  org.CreatedAt,
		Groups:     posthog.NewGroups().Set("organization", org.ID),
		Properties: posthog.NewProperties().
			Set("org_id", org.ID).
			Set("org_name", org.Name).
			Set("org_type", string(org.OrgType)).
			Set("tags", tags).
			Set("source", source).
			Set("is_first_org", len(account.OrgIDs) == 1).
			Set("$set", personSet),
	}

	return groupIdentify, capture
}
