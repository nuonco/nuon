package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
)

type ListAppsNeedingDefaultBranchRequest struct {
	// OrgIDs narrows the backfill to a subset of orgs. Empty covers every org.
	OrgIDs []string `json:"org_ids"`
}

type ListAppsNeedingDefaultBranchResponse struct {
	AppIDs []string `json:"app_ids"`
}

// ListAppsNeedingDefaultBranch returns the apps that have no branch named
// `default`, plus the ones that have it but still hold installs on no branch at
// all, oldest first. The second case covers an app whose branch the CLI created
// before the backfill existed, whose installs were never recorded as members.
// Apps on their way out are left alone: a branch and its queues would outlive
// the app they were created for.
//
// @temporal-gen-v2 activity
// @start-to-close-timeout 2m
func (a *Activities) ListAppsNeedingDefaultBranch(ctx context.Context, req ListAppsNeedingDefaultBranchRequest) (*ListAppsNeedingDefaultBranchResponse, error) {
	query := a.db.WithContext(ctx).
		Model(&app.App{}).
		Select("id").
		Where("status IS NULL OR status NOT IN ?", []app.AppStatus{
			app.AppStatusDeprovisioning,
			app.AppStatusDeleteQueued,
		})
	if len(req.OrgIDs) > 0 {
		query = query.Where("org_id IN ?", req.OrgIDs)
	}

	var apps []app.App
	if err := query.Order("created_at ASC").Find(&apps).Error; err != nil {
		return nil, fmt.Errorf("unable to list apps: %w", err)
	}

	var branches []app.AppBranch
	if err := a.db.WithContext(ctx).
		Model(&app.AppBranch{}).
		Select("app_id").
		Where(app.AppBranch{Name: appshelpers.DefaultAppBranchName}).
		Find(&branches).Error; err != nil {
		return nil, fmt.Errorf("unable to list default app branches: %w", err)
	}

	haveDefault := make(map[string]struct{}, len(branches))
	for _, branch := range branches {
		haveDefault[branch.AppID] = struct{}{}
	}

	var appIDsWithLooseInstalls []string
	if err := a.db.WithContext(ctx).
		Model(&app.Install{}).
		Distinct("app_id").
		Where("app_branch_id IS NULL").
		Pluck("app_id", &appIDsWithLooseInstalls).Error; err != nil {
		return nil, fmt.Errorf("unable to list apps with unconnected installs: %w", err)
	}

	looseInstalls := make(map[string]struct{}, len(appIDsWithLooseInstalls))
	for _, appID := range appIDsWithLooseInstalls {
		looseInstalls[appID] = struct{}{}
	}

	resp := &ListAppsNeedingDefaultBranchResponse{AppIDs: make([]string, 0, len(apps))}
	for _, ap := range apps {
		_, hasBranch := haveDefault[ap.ID]
		_, hasLooseInstalls := looseInstalls[ap.ID]
		if hasBranch && !hasLooseInstalls {
			continue
		}
		resp.AppIDs = append(resp.AppIDs, ap.ID)
	}

	return resp, nil
}
