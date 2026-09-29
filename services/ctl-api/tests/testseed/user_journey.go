package testseed

import (
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

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
