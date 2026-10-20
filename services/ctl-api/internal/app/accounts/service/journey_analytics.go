package service

import (
	"slices"

	"github.com/gin-gonic/gin"
	posthog "github.com/posthog/posthog-go"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const (
	journeyStepCompleteEvent = "journey_step_complete"
	journeyCompleteEvent     = "journey_complete"
	journeyPathChosenEvent   = "journey_path_chosen"
)

var journeyAnalyticsMetadataKeys = []string{"path", "cloud"}

func (s *service) trackJourneyEvent(ctx *gin.Context, account *app.Account, event string, journey app.UserJourney, stepName string) {
	switch account.AccountType {
	case app.AccountTypeService, app.AccountTypeCanary, app.AccountTypeIntegration:
		return
	}

	distinctID := account.Email
	if distinctID == "" {
		distinctID = account.ID
	}

	orgID := ctx.GetHeader("X-Nuon-Org-ID")
	if !slices.Contains(account.OrgIDs, orgID) {
		orgID = ""
	}
	if orgID == "" && len(account.OrgIDs) == 1 {
		orgID = account.OrgIDs[0]
	}

	props := posthog.NewProperties().Set("journey", journey.Name)
	if stepName != "" {
		props.Set("step", stepName)
	}
	for _, step := range journey.Steps {
		for _, key := range journeyAnalyticsMetadataKeys {
			if v, ok := step.Metadata[key].(string); ok && v != "" {
				props.Set(key, v)
			}
		}
	}
	if orgID != "" {
		props.Set("org_id", orgID)
	}

	s.productAnalytics.Capture(distinctID, event, orgID, props)
}
