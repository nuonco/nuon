package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type EnsureOrgQueueRequest struct {
	OrgID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field OrgID
func (a *Activities) EnsureOrgQueue(ctx context.Context, req EnsureOrgQueueRequest) error {
	return a.helpers.EnsureOrgQueue(ctx, req.OrgID)
}

type EnsureInstallQueuesRequest struct {
	InstallID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field InstallID
func (a *Activities) EnsureInstallQueues(ctx context.Context, req EnsureInstallQueuesRequest) error {
	return a.installsHelpers.EnsureInstallQueues(ctx, req.InstallID)
}

type EnsureAppQueueRequest struct {
	AppID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field AppID
func (a *Activities) EnsureAppQueue(ctx context.Context, req EnsureAppQueueRequest) error {
	return a.appsHelpers.EnsureAppQueue(ctx, req.AppID)
}

type EnsureAppBranchQueueRequest struct {
	BranchID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field BranchID
func (a *Activities) EnsureAppBranchQueue(ctx context.Context, req EnsureAppBranchQueueRequest) error {
	return a.appsHelpers.EnsureAppBranchQueue(ctx, req.BranchID)
}

type EnsureComponentQueueRequest struct {
	ComponentID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field ComponentID
func (a *Activities) EnsureComponentQueue(ctx context.Context, req EnsureComponentQueueRequest) error {
	_, err := a.componentHelpers.EnsureComponentQueues(ctx, req.ComponentID)
	return err
}

type EnsureRunnerQueuesRequest struct {
	RunnerID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field RunnerID
func (a *Activities) EnsureRunnerQueues(ctx context.Context, req EnsureRunnerQueuesRequest) error {
	var runner app.Runner
	if res := a.db.WithContext(ctx).
		Preload("RunnerGroup").
		Where(app.Runner{ID: req.RunnerID}).
		First(&runner); res.Error != nil {
		return fmt.Errorf("unable to get runner: %w", res.Error)
	}

	if err := a.runnersHelpers.EnsureRunnerSignalsQueue(ctx, runner.ID); err != nil {
		return fmt.Errorf("unable to ensure runner signals queue: %w", err)
	}

	if err := a.runnersHelpers.EnsureRunnerJobGroupQueues(ctx, &runner, &runner.RunnerGroup.Settings); err != nil {
		return fmt.Errorf("unable to ensure runner job group queues: %w", err)
	}

	return nil
}

type EnsureVCSConnectionQueueRequest struct {
	VCSConnectionID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field VCSConnectionID
func (a *Activities) EnsureVCSConnectionQueue(ctx context.Context, req EnsureVCSConnectionQueueRequest) error {
	var vcsConn app.VCSConnection
	if res := a.db.WithContext(ctx).
		Where(app.VCSConnection{ID: req.VCSConnectionID}).
		First(&vcsConn); res.Error != nil {
		return fmt.Errorf("unable to get vcs connection: %w", res.Error)
	}

	return a.vcsHelpers.EnsureConnectionQueue(ctx, &vcsConn)
}

type GetInstallRunnersRequest struct {
	InstallID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field InstallID
func (a *Activities) GetInstallRunners(ctx context.Context, req GetInstallRunnersRequest) ([]app.Runner, error) {
	var rg app.RunnerGroup
	if res := a.db.WithContext(ctx).
		Where(app.RunnerGroup{OwnerID: req.InstallID, OwnerType: "installs"}).
		First(&rg); res.Error != nil {
		return nil, fmt.Errorf("unable to get install runner group: %w", res.Error)
	}

	var runners []app.Runner
	if res := a.db.WithContext(ctx).
		Where(app.Runner{RunnerGroupID: rg.ID}).
		Find(&runners); res.Error != nil {
		return nil, fmt.Errorf("unable to get install runners: %w", res.Error)
	}
	return runners, nil
}

type GetOrgAppsRequest struct {
	OrgID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field OrgID
func (a *Activities) GetOrgApps(ctx context.Context, req GetOrgAppsRequest) ([]app.App, error) {
	var apps []app.App
	if res := a.db.WithContext(ctx).
		Where(app.App{OrgID: req.OrgID}).
		Find(&apps); res.Error != nil {
		return nil, fmt.Errorf("unable to get org apps: %w", res.Error)
	}
	return apps, nil
}

type GetAppBranchesRequest struct {
	AppID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field AppID
func (a *Activities) GetAppBranches(ctx context.Context, req GetAppBranchesRequest) ([]app.AppBranch, error) {
	var branches []app.AppBranch
	if res := a.db.WithContext(ctx).
		Where(app.AppBranch{AppID: req.AppID}).
		Find(&branches); res.Error != nil {
		return nil, fmt.Errorf("unable to get app branches: %w", res.Error)
	}
	return branches, nil
}

type GetAppComponentsRequest struct {
	AppID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field AppID
func (a *Activities) GetAppComponents(ctx context.Context, req GetAppComponentsRequest) ([]app.Component, error) {
	var components []app.Component
	if res := a.db.WithContext(ctx).
		Where(app.Component{AppID: req.AppID}).
		Find(&components); res.Error != nil {
		return nil, fmt.Errorf("unable to get app components: %w", res.Error)
	}
	return components, nil
}

type GetOrgInstallsRequest struct {
	OrgID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field OrgID
func (a *Activities) GetOrgInstalls(ctx context.Context, req GetOrgInstallsRequest) ([]app.Install, error) {
	var installs []app.Install
	if res := a.db.WithContext(ctx).
		Where(app.Install{OrgID: req.OrgID}).
		Find(&installs); res.Error != nil {
		return nil, fmt.Errorf("unable to get org installs: %w", res.Error)
	}
	return installs, nil
}

type GetOrgVCSConnectionsRequest struct {
	OrgID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field OrgID
func (a *Activities) GetOrgVCSConnections(ctx context.Context, req GetOrgVCSConnectionsRequest) ([]app.VCSConnection, error) {
	var conns []app.VCSConnection
	if res := a.db.WithContext(ctx).
		Where(app.VCSConnection{OrgID: req.OrgID}).
		Find(&conns); res.Error != nil {
		return nil, fmt.Errorf("unable to get org vcs connections: %w", res.Error)
	}
	return conns, nil
}

type EnableQueuesFeatureFlagRequest struct {
	OrgID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field OrgID
func (a *Activities) EnableQueuesFeatureFlag(ctx context.Context, req EnableQueuesFeatureFlagRequest) error {
	return a.features.Enable(ctx, req.OrgID, map[string]bool{
		string(app.OrgFeatureQueues):             true,
		string(app.OrgFeatureAppBranches):        true,
		string(app.OrgFeatureParallelRunnerJobs): true,
	})
}

type UpdateOrgStatusV2MetadataRequest struct {
	OrgID string         `validate:"required"`
	Data  map[string]any `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) UpdateOrgStatusV2Metadata(ctx context.Context, req UpdateOrgStatusV2MetadataRequest) error {
	var org app.Org
	if res := a.db.WithContext(ctx).Select("id", "status_v2").First(&org, "id = ?", req.OrgID); res.Error != nil {
		return fmt.Errorf("unable to get org: %w", res.Error)
	}

	if org.StatusV2.Metadata == nil {
		org.StatusV2.Metadata = make(map[string]any)
	}
	for k, v := range req.Data {
		org.StatusV2.Metadata[k] = v
	}

	if res := a.db.WithContext(ctx).Model(&org).Select("status_v2").Updates(&org); res.Error != nil {
		return fmt.Errorf("unable to update org status_v2 metadata: %w", res.Error)
	}

	return nil
}
