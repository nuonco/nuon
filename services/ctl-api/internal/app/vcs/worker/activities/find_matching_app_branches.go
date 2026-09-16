package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
)

type FindMatchingAppBranchesRequest struct {
	OrgID  string `json:"org_id" validate:"required"`
	Repo   string `json:"repo" validate:"required"`
	Branch string `json:"branch,omitempty"`
}

type MatchingAppBranch struct {
	AppBranchID       string                 `json:"app_branch_id"`
	AppBranchConfigID string                 `json:"app_branch_config_id"`
	RunConfig         app.AppBranchRunConfig `json:"run_config" gorm:"-"`
}

// @temporal-gen-v2 activity
func (a *Activities) FindMatchingAppBranches(ctx context.Context, req FindMatchingAppBranchesRequest) ([]MatchingAppBranch, error) {
	appBranchConfigTable := plugins.TableName(a.db, app.AppBranchConfig{})

	results, err := a.findMatchesByConfigType(ctx, req, appBranchConfigTable)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	var deduped []MatchingAppBranch
	for _, r := range results {
		if seen[r.AppBranchID] {
			continue
		}
		seen[r.AppBranchID] = true
		deduped = append(deduped, r)
	}

	configIDs := make([]string, 0, len(deduped))
	for _, match := range deduped {
		configIDs = append(configIDs, match.AppBranchConfigID)
	}
	var configs []app.AppBranchConfig
	if len(configIDs) > 0 {
		if err := a.db.WithContext(ctx).
			Select("id", "run_config").
			Where("id IN ?", configIDs).
			Find(&configs).Error; err != nil {
			return nil, fmt.Errorf("unable to load app branch run configs: %w", err)
		}
	}
	runConfigs := make(map[string]app.AppBranchRunConfig, len(configs))
	for _, config := range configs {
		runConfig := app.AppBranchRunConfig{Mode: app.AppBranchRunModePush}
		if config.RunConfig != nil {
			runConfig = *config.RunConfig
			runConfig.Normalize()
		}
		runConfigs[config.ID] = runConfig
	}
	for i := range deduped {
		runConfig, ok := runConfigs[deduped[i].AppBranchConfigID]
		if !ok {
			return nil, fmt.Errorf("run config not loaded for app branch config %s", deduped[i].AppBranchConfigID)
		}
		deduped[i].RunConfig = runConfig
	}

	return deduped, nil
}

func (a *Activities) findMatchesByConfigType(ctx context.Context, req FindMatchingAppBranchesRequest, configType string) ([]MatchingAppBranch, error) {
	var results []MatchingAppBranch

	connectedQuery := a.db.WithContext(ctx).
		Table("connected_github_vcs_configs").
		Select("app_branch_configs.app_branch_id, app_branch_configs.id as app_branch_config_id").
		Joins("JOIN app_branch_configs ON app_branch_configs.id = connected_github_vcs_configs.component_config_id AND connected_github_vcs_configs.component_config_type = ?", configType).
		Joins("JOIN app_branches ON app_branches.id = app_branch_configs.app_branch_id AND app_branches.deleted_at = 0").
		Where("connected_github_vcs_configs.org_id = ?", req.OrgID).
		Where("connected_github_vcs_configs.repo = ?", req.Repo).
		Where("connected_github_vcs_configs.deleted_at = 0").
		Where("app_branch_configs.deleted_at = 0").
		Order("app_branch_configs.created_at DESC")
	if req.Branch != "" {
		connectedQuery = connectedQuery.Where("connected_github_vcs_configs.branch = ?", req.Branch)
	}
	err := connectedQuery.Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("unable to find matching app branches (connected): %w", err)
	}

	var publicResults []MatchingAppBranch
	publicQuery := a.db.WithContext(ctx).
		Table("public_git_vcs_configs").
		Select("app_branch_configs.app_branch_id, app_branch_configs.id as app_branch_config_id").
		Joins("JOIN app_branch_configs ON app_branch_configs.id = public_git_vcs_configs.component_config_id AND public_git_vcs_configs.component_config_type = ?", configType).
		Joins("JOIN app_branches ON app_branches.id = app_branch_configs.app_branch_id AND app_branches.deleted_at = 0").
		Where("app_branch_configs.org_id = ?", req.OrgID).
		Where("(public_git_vcs_configs.repo = ? OR public_git_vcs_configs.repo = ?)", req.Repo, "https://github.com/"+req.Repo+".git").
		Where("public_git_vcs_configs.deleted_at = 0").
		Where("app_branch_configs.deleted_at = 0").
		Order("app_branch_configs.created_at DESC")
	if req.Branch != "" {
		publicQuery = publicQuery.Where("public_git_vcs_configs.branch = ?", req.Branch)
	}
	err = publicQuery.Scan(&publicResults).Error
	if err != nil {
		return nil, fmt.Errorf("unable to find matching app branches (public): %w", err)
	}

	return append(results, publicResults...), nil
}
