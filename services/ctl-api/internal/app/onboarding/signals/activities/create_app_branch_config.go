package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 2m
// @as-wrapper
func (a *Activities) createOnboardingAppBranchConfig(ctx context.Context, appBranchID, repo, directory, branch string) (*app.AppBranchConfig, error) {
	config, err := a.appsHelpers.CreateAppBranchConfig(ctx, appBranchID,
		nil,
		&app.PublicGitVCSConfig{
			Repo:      repo,
			Directory: directory,
			Branch:    branch,
		},
		nil,
		nil,
		nil,
		nil,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create app branch config: %w", err)
	}

	if err := a.appsHelpers.EnqueueAppBranchConfigSignals(ctx, appBranchID, config.ID); err != nil {
		return nil, fmt.Errorf("unable to enqueue app branch config signals: %w", err)
	}

	return config, nil
}
