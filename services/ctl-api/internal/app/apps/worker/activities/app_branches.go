package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type ParseNuonYamlRequest struct {
	VCSConfigID string
	CommitSHA   string
}

type CreateAppConfigRequest struct {
	AppID  string
	Config interface{} // TODO: Define proper type based on parsed nuon.yaml structure
}

// @temporal-gen as-activity
func (a *Activities) GetAppBranchByID(ctx context.Context, appBranchID string) (*app.AppBranch, error) {
	var branch app.AppBranch
	if err := a.db.WithContext(ctx).
		Preload("ConnectedGithubVCSConfig").
		Preload("Queue").
		First(&branch, "id = ?", appBranchID).Error; err != nil {
		return nil, fmt.Errorf("unable to get app branch: %w", err)
	}
	return &branch, nil
}

// @temporal-gen as-activity
func (a *Activities) GetLatestCommitFromVCS(ctx context.Context, vcsConfigID string) (string, error) {
	// TODO: Implement VCS API call to get latest commit SHA
	// This will need to:
	// 1. Get the VCS config
	// 2. Make GitHub API call to get latest commit
	// 3. Return commit SHA
	return "", fmt.Errorf("not implemented: GetLatestCommitFromVCS")
}

// @temporal-gen as-activity
func (a *Activities) ParseNuonYamlFromRepo(ctx context.Context, req ParseNuonYamlRequest) (interface{}, error) {
	// TODO: Implement repo cloning and nuon.yaml parsing
	// This will need to:
	// 1. Clone or fetch the repo at the specific commit
	// 2. Read nuon.yaml file
	// 3. Parse using existing nuon config parser
	// 4. Return parsed config structure
	return nil, fmt.Errorf("not implemented: ParseNuonYamlFromRepo")
}

// @temporal-gen as-activity
func (a *Activities) CreateAppConfig(ctx context.Context, req CreateAppConfigRequest) (*app.AppConfig, error) {
	// TODO: Implement app config creation from parsed nuon.yaml
	// This will need to:
	// 1. Create AppConfig record
	// 2. Create related records (components, etc.)
	// 3. Return created AppConfig
	return nil, fmt.Errorf("not implemented: CreateAppConfig")
}

// @temporal-gen as-activity
func (a *Activities) UpdateAppBranchLastSyncedCommit(ctx context.Context, appBranchID, commitSHA string) error {
	return a.db.WithContext(ctx).
		Model(&app.AppBranch{}).
		Where("id = ?", appBranchID).
		Update("last_synced_commit", commitSHA).Error
}

// @temporal-gen as-activity
func (a *Activities) GetAppConfigByID(ctx context.Context, appConfigID string) (*app.AppConfig, error) {
	var config app.AppConfig
	if err := a.db.WithContext(ctx).
		Preload("Components").
		First(&config, "id = ?", appConfigID).Error; err != nil {
		return nil, fmt.Errorf("unable to get app config: %w", err)
	}
	return &config, nil
}

// @temporal-gen as-activity
func (a *Activities) TriggerComponentBuild(ctx context.Context, componentID string) error {
	// TODO: Implement component build trigger
	// This will need to:
	// 1. Create a component build record
	// 2. Enqueue signal to component queue for building
	// 3. Return immediately (async build)
	return fmt.Errorf("not implemented: TriggerComponentBuild")
}
