package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/posthog"
)

const orgCreatedEvent = "org_created"

type SendOrgCreatedEventRequest struct {
	OrgID  string `json:"org_id" validate:"required"`
	Source string `json:"source"`
}

// @temporal-gen-v2 activity
func (a *Activities) SendOrgCreatedEvent(ctx context.Context, req SendOrgCreatedEventRequest) (bool, error) {
	if !a.posthog.Enabled() {
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

	if err := a.posthog.Capture(ctx, buildOrgCreatedEvents(&account, org, req.Source)...); err != nil {
		return false, fmt.Errorf("unable to send org created event: %w", err)
	}

	return true, nil
}

func buildOrgCreatedEvents(account *app.Account, org *app.Org, source string) []posthog.Event {
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

	return []posthog.Event{
		{
			Event:      "$groupidentify",
			DistinctID: distinctID,
			Timestamp:  org.CreatedAt,
			Properties: map[string]any{
				"$group_type": "organization",
				"$group_key":  org.ID,
				"$group_set": map[string]any{
					"name":       org.Name,
					"org_type":   string(org.OrgType),
					"tags":       org.Tags,
					"created_at": org.CreatedAt,
				},
			},
		},
		{
			Event:      orgCreatedEvent,
			DistinctID: distinctID,
			Timestamp:  org.CreatedAt,
			Properties: map[string]any{
				"org_id":       org.ID,
				"org_name":     org.Name,
				"org_type":     string(org.OrgType),
				"tags":         org.Tags,
				"source":       source,
				"is_first_org": len(account.OrgIDs) == 1,
				"$groups":      map[string]any{"organization": org.ID},
				"$set":         personSet,
			},
		},
	}
}
