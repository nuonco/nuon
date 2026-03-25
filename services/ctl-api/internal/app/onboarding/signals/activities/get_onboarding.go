package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
// @as-wrapper
// @by-field onboardingID
func (a *Activities) getOnboarding(ctx context.Context, onboardingID string) (*app.Onboarding, error) {
	var onboarding app.Onboarding
	res := a.db.WithContext(ctx).First(&onboarding, "id = ?", onboardingID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get onboarding: %w", res.Error)
	}
	return &onboarding, nil
}
