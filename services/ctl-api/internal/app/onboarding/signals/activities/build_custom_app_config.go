package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/pkg/config/builder"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 2m
// @as-wrapper
func (a *Activities) buildCustomAppConfig(ctx context.Context, onboardingID string) (*app.Onboarding, error) {
	var onboarding app.Onboarding
	if err := a.db.WithContext(ctx).First(&onboarding, "id = ?", onboardingID).Error; err != nil {
		return nil, fmt.Errorf("unable to get onboarding: %w", err)
	}

	if onboarding.CloudProvider == nil || *onboarding.CloudProvider == "" {
		return nil, fmt.Errorf("onboarding has no cloud_provider set")
	}

	var attrs []string
	if len(onboarding.AppAttributes) > 0 {
		attrs = []string(onboarding.AppAttributes)
	}

	b, err := builder.New(*onboarding.CloudProvider)
	if err != nil {
		return nil, fmt.Errorf("unable to create config builder: %w", err)
	}

	cfg, err := b.Build(attrs)
	if err != nil {
		return nil, fmt.Errorf("unable to build app config: %w", err)
	}

	onboarding.AppConfig = &app.IntermediateAppConfig{AppConfig: *cfg}

	if err := a.db.WithContext(ctx).Save(&onboarding).Error; err != nil {
		return nil, fmt.Errorf("unable to save onboarding with app config: %w", err)
	}

	return &onboarding, nil
}
