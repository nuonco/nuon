package activities

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

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
// of every app also being a migration, and connects the app's existing installs
// to that branch.
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
			resp.InstallsConnected, err = a.connectInstallsToBranch(ctx, ap.ID, defaultBranch.ID)
			if err != nil {
				return nil, err
			}
			return resp, nil
		}
	}

	// Two branches claiming all installs would both deploy to the same installs, so
	// an app already routing everything through a differently-named branch is left
	// alone rather than given a competing default.
	claimed, err := a.branchClaimsAllInstalls(ctx, branches, appshelpers.DefaultAppBranchName)
	if err != nil {
		return nil, err
	}
	if claimed {
		resp.Outcome = DefaultAppBranchOutcomeClaimed
		return resp, nil
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

	if _, err := a.appsHelpers.CreateAppBranchConfig(
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
	); err != nil {
		return nil, fmt.Errorf("unable to configure default branch %s: %w", defaultBranch.ID, err)
	}

	connected, err := a.connectInstallsToBranch(ctx, ap.ID, defaultBranch.ID)
	if err != nil {
		return nil, err
	}

	resp.Outcome = DefaultAppBranchOutcomeCreated
	resp.InstallsConnected = connected
	return resp, nil
}

// connectInstallsToBranch records the installs the branch's all-installs group
// already deploys to as members of the branch, so an install shows the branch it
// belongs to rather than none at all. It claims exactly what the group resolves,
// which leaves installs owned by another branch alone.
func (a *Activities) connectInstallsToBranch(ctx context.Context, appID, branchID string) (int, error) {
	installIDs, err := a.appsHelpers.ResolveAllInstallsForBranch(ctx, appID, branchID)
	if err != nil {
		return 0, fmt.Errorf("unable to resolve installs for branch %s: %w", branchID, err)
	}
	if len(installIDs) == 0 {
		return 0, nil
	}

	var unconnected []app.Install
	if err := a.db.WithContext(ctx).
		Where("id IN ?", installIDs).
		Where("app_branch_id IS NULL").
		Find(&unconnected).Error; err != nil {
		return 0, fmt.Errorf("unable to load installs for branch %s: %w", branchID, err)
	}

	for i := range unconnected {
		a.appsHelpers.SyncInstallBranchConnection(ctx, &unconnected[i], branchID)
	}

	if err := a.verifyInstallsConnected(ctx, installIDs, branchID); err != nil {
		return 0, err
	}

	return len(unconnected), nil
}

// verifyInstallsConnected re-reads both halves of branch membership, because
// SyncInstallBranchConnection drops its write errors on the floor and a
// migration must not report a count that nothing wrote. The connection row is
// what the dashboard reads; the pin is what the branch validators read.
func (a *Activities) verifyInstallsConnected(ctx context.Context, installIDs []string, branchID string) error {
	var pinned int64
	if err := a.db.WithContext(ctx).
		Model(&app.Install{}).
		Where("id IN ?", installIDs).
		Where("app_branch_id = ?", branchID).
		Count(&pinned).Error; err != nil {
		return fmt.Errorf("unable to verify install pins for branch %s: %w", branchID, err)
	}
	if int(pinned) < len(installIDs) {
		return fmt.Errorf("pinned %d of %d installs to branch %s", pinned, len(installIDs), branchID)
	}

	var connected int64
	if err := a.db.WithContext(ctx).
		Model(&app.InstallAppBranchConnection{}).
		Where(app.InstallAppBranchConnection{AppBranchID: branchID, Active: true}).
		Where("install_id IN ?", installIDs).
		Count(&connected).Error; err != nil {
		return fmt.Errorf("unable to verify install connections for branch %s: %w", branchID, err)
	}
	if int(connected) < len(installIDs) {
		return fmt.Errorf("connected %d of %d installs to branch %s", connected, len(installIDs), branchID)
	}

	return nil
}

func (a *Activities) branchClaimsAllInstalls(ctx context.Context, branches []app.AppBranch, excludeName string) (bool, error) {
	for _, branch := range branches {
		if branch.Name == excludeName {
			continue
		}

		var latest app.AppBranchConfig
		err := a.db.WithContext(ctx).
			Where(app.AppBranchConfig{AppBranchID: branch.ID}).
			Order("created_at DESC").
			First(&latest).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			continue
		}
		if err != nil {
			return false, fmt.Errorf("unable to load latest config for branch %s: %w", branch.ID, err)
		}

		var count int64
		if err := a.db.WithContext(ctx).
			Model(&app.AppBranchInstallGroup{}).
			Where(app.AppBranchInstallGroup{AppBranchConfigID: latest.ID, AllInstalls: true}).
			Count(&count).Error; err != nil {
			return false, fmt.Errorf("unable to count all-installs groups for config %s: %w", latest.ID, err)
		}
		if count > 0 {
			return true, nil
		}
	}

	return false, nil
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
