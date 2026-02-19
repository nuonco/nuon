package testseed

import (
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// BuildUserJourney creates an app.UserJourney with the default onboarding steps,
// all marked incomplete. Override fields directly when you need custom values.
func BuildUserJourney() app.UserJourney {
	return app.UserJourney{
		Name:  "onboarding",
		Title: "Getting Started",
		Steps: []app.UserJourneyStep{
			{Name: "create-org", Title: "Create Organization", Complete: false},
			{Name: "create-app", Title: "Create App", Complete: false},
			{Name: "create-install", Title: "Create Install", Complete: false},
		},
	}
}

// BuildCompletedUserJourney creates an app.UserJourney with the default onboarding steps,
// all marked complete. Override fields directly when you need custom values.
func BuildCompletedUserJourney() app.UserJourney {
	now := time.Now()
	return app.UserJourney{
		Name:  "onboarding",
		Title: "Getting Started",
		Steps: []app.UserJourneyStep{
			{Name: "create-org", Title: "Create Organization", Complete: true, CompletedAt: &now, CompletionMethod: "test", CompletionSource: "testseed"},
			{Name: "create-app", Title: "Create App", Complete: true, CompletedAt: &now, CompletionMethod: "test", CompletionSource: "testseed"},
			{Name: "create-install", Title: "Create Install", Complete: true, CompletedAt: &now, CompletionMethod: "test", CompletionSource: "testseed"},
		},
	}
}
