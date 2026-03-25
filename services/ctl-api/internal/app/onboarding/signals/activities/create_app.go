package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appsignals "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type CreateOnboardingAppResponse struct {
	AppID       string `json:"app_id"`
	AppBranchID string `json:"app_branch_id"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 5m
// @as-wrapper
func (a *Activities) createOnboardingApp(ctx context.Context, orgID, appName string) (*CreateOnboardingAppResponse, error) {
	// Load org for context — needed for event loop startup (startEventLoop calls signal.GetOrg)
	var org app.Org
	if err := a.db.WithContext(ctx).First(&org, "id = ?", orgID).Error; err != nil {
		return nil, fmt.Errorf("unable to get org: %w", err)
	}
	ctx = cctx.SetOrgContext(ctx, &org)

	// Idempotency: if app already exists for this org+name, reuse it
	var existingApp app.App
	if err := a.db.WithContext(ctx).
		Preload("AppBranches").
		Where("org_id = ? AND name = ?", orgID, appName).
		First(&existingApp).Error; err == nil {
		branchID := ""
		if len(existingApp.AppBranches) > 0 {
			branchID = existingApp.AppBranches[0].ID
			// Ensure branch queue exists (may have been missed in a previous failed attempt)
			if err := a.appsHelpers.EnsureAppBranchQueue(ctx, branchID); err != nil {
				return nil, fmt.Errorf("unable to ensure app branch queue: %w", err)
			}
		}
		return &CreateOnboardingAppResponse{
			AppID:       existingApp.ID,
			AppBranchID: branchID,
		}, nil
	}

	// Create the app record
	newApp := app.App{
		OrgID:             orgID,
		Name:              appName,
		Status:            "queued",
		StatusDescription: "waiting for event loop to start and provision app",
		DisplayName:       generics.NewNullString(appName),
		NotificationsConfig: app.NotificationsConfig{
			EnableSlackNotifications: true,
			EnableEmailNotifications: true,
		},
	}

	if err := a.db.WithContext(ctx).Create(&newApp).Error; err != nil {
		return nil, fmt.Errorf("unable to create app: %w", err)
	}

	// Create sandbox queue for the app
	if err := a.appsHelpers.CreateAppSandboxQueue(ctx, newApp.ID); err != nil {
		return nil, fmt.Errorf("unable to create app sandbox queue: %w", err)
	}

	// Create the main branch + branch queue
	branch, err := a.appsHelpers.CreateAppBranch(ctx, newApp.ID, "main")
	if err != nil {
		return nil, fmt.Errorf("unable to create app branch: %w", err)
	}

	// Send v1 event loop signals (matching apps/service/create_app.go)
	a.evClient.Send(ctx, newApp.ID, &appsignals.Signal{
		Type: appsignals.OperationCreated,
	})
	a.evClient.Send(ctx, newApp.ID, &appsignals.Signal{
		Type: appsignals.OperationPollDependencies,
	})
	a.evClient.Send(ctx, newApp.ID, &appsignals.Signal{
		Type: appsignals.OperationProvision,
	})

	return &CreateOnboardingAppResponse{
		AppID:       newApp.ID,
		AppBranchID: branch.ID,
	}, nil
}
