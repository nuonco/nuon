package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

const (
	DefaultAppBranchOutcomeCreated = "created"
	DefaultAppBranchOutcomeExists  = "exists"
	DefaultAppBranchOutcomeClaimed = "claimed_by_other_branch"
)

type EnsureDefaultAppBranchRequest struct {
	AppID string `json:"app_id"`
}

type EnsureDefaultAppBranchResponse struct {
	AppID             string `json:"app_id"`
	BranchID          string `json:"branch_id,omitempty"`
	Outcome           string `json:"outcome"`
	InstallsConnected int    `json:"installs_connected"`
}

// EnsureDefaultAppBranch gives one app the `default` branch and single
// all-installs group that `nuon apps sync` creates on its first run once
// default-app-branches is on, so the flag can be flipped without the first sync
// of every app also creating the branch. Existing installs remain unbranched;
// ownership is only selected during install creation or an explicit move.
//
// Idempotent in stages rather than transactionally: a retry after the branch
// landed but its config did not finds the branch and only adds the config, and
// queue creation starts Temporal workflows so it cannot share a transaction
// with the branch insert anyway.
//
// @temporal-gen-v2 activity
// @start-to-close-timeout 2m
func (a *Activities) EnsureDefaultAppBranch(ctx context.Context, req EnsureDefaultAppBranchRequest) (*EnsureDefaultAppBranchResponse, error) {
	var ap app.App
	if err := a.db.WithContext(ctx).
		Select("id", "org_id", "created_by_id").
		First(&ap, "id = ?", req.AppID).Error; err != nil {
		return nil, fmt.Errorf("unable to get app %s: %w", req.AppID, err)
	}

	// The branch, its config and its install group all have NOT NULL org_id and
	// created_by_id filled from context. An activity has no account, so attribute
	// the rows to whoever created the app, the way the phone-home backfill
	// attributes to the org's creator.
	ctx = cctx.SetOrgIDContext(ctx, ap.OrgID)
	ctx = cctx.SetAccountIDContext(ctx, ap.CreatedByID)

	var branches []app.AppBranch
	if err := a.db.WithContext(ctx).
		Where(app.AppBranch{AppID: ap.ID}).
		Find(&branches).Error; err != nil {
		return nil, fmt.Errorf("unable to list branches for app %s: %w", ap.ID, err)
	}

	resp := &EnsureDefaultAppBranchResponse{AppID: ap.ID}

	var defaultBranch *app.AppBranch
	for i := range branches {
		if branches[i].Name == appshelpers.DefaultAppBranchName {
			defaultBranch = &branches[i]
			break
		}
	}

	if defaultBranch != nil {
		resp.BranchID = defaultBranch.ID

		hasConfig, err := a.branchHasConfig(ctx, defaultBranch.ID)
		if err != nil {
			return nil, err
		}
		if hasConfig {
			resp.Outcome = DefaultAppBranchOutcomeExists
			return resp, nil
		}
	}

	if defaultBranch == nil {
		branch, err := a.appsHelpers.CreateAppBranch(ctx, ap.ID, appshelpers.DefaultAppBranchName)
		if err != nil {
			return nil, fmt.Errorf("unable to create default branch for app %s: %w", ap.ID, err)
		}
		defaultBranch = branch
		resp.BranchID = branch.ID
	} else if err := a.appsHelpers.EnsureAppBranchQueues(ctx, defaultBranch.ID); err != nil {
		return nil, fmt.Errorf("unable to ensure queues for branch %s: %w", defaultBranch.ID, err)
	}

	config, err := a.appsHelpers.CreateAppBranchConfig(
		ctx,
		defaultBranch.ID,
		nil,
		nil,
		[]app.AppBranchInstallGroup{{
			Name:        appshelpers.DefaultAppBranchInstallGroupName,
			Order:       0,
			AllInstalls: true,
		}},
		&[]string{},
		nil,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to configure default branch %s: %w", defaultBranch.ID, err)
	}
	if err := a.appsHelpers.EnqueueAppBranchCreatedIfFirst(ctx, defaultBranch.ID, config.ID); err != nil {
		return nil, fmt.Errorf("unable to enqueue app-branch-created for %s: %w", defaultBranch.ID, err)
	}

	resp.Outcome = DefaultAppBranchOutcomeCreated
	return resp, nil
}

func (a *Activities) branchHasConfig(ctx context.Context, branchID string) (bool, error) {
	var count int64
	if err := a.db.WithContext(ctx).
		Model(&app.AppBranchConfig{}).
		Where(app.AppBranchConfig{AppBranchID: branchID}).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("unable to count configs for branch %s: %w", branchID, err)
	}
	return count > 0, nil
}
