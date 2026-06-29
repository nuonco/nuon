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
	Branch string `json:"branch" validate:"required"`
}

type MatchingAppBranch struct {
	AppBranchID       string `json:"app_branch_id"`
	AppBranchConfigID string `json:"app_branch_config_id"`
}

// @temporal-gen-v2 activity
func (a *Activities) FindMatchingAppBranches(ctx context.Context, req FindMatchingAppBranchesRequest) ([]MatchingAppBranch, error) {
	appBranchConfigTable := plugins.TableName(a.db, app.AppBranchConfig{})

	var results []MatchingAppBranch

	// Match via connected_github_vcs_configs.
	err := a.db.WithContext(ctx).
		Table("connected_github_vcs_configs").
		Select("app_branch_configs.app_branch_id, app_branch_configs.id as app_branch_config_id").
		Joins("JOIN app_branch_configs ON app_branch_configs.id = connected_github_vcs_configs.component_config_id AND connected_github_vcs_configs.component_config_type = ?", appBranchConfigTable).
		Where("connected_github_vcs_configs.org_id = ?", req.OrgID).
		Where("connected_github_vcs_configs.repo = ?", req.Repo).
		Where("connected_github_vcs_configs.branch = ?", req.Branch).
		Where("connected_github_vcs_configs.deleted_at = 0").
		Where("app_branch_configs.deleted_at = 0").
		Order("app_branch_configs.created_at DESC").
		Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("unable to find matching app branches (connected): %w", err)
	}

	// Match via public_git_vcs_configs.
	// Public git configs store repo as a clone URL (e.g. "https://github.com/owner/repo.git")
	// while GitHub webhook payloads provide "owner/repo". Match both formats.
	var publicResults []MatchingAppBranch
	err = a.db.WithContext(ctx).
		Table("public_git_vcs_configs").
		Select("app_branch_configs.app_branch_id, app_branch_configs.id as app_branch_config_id").
		Joins("JOIN app_branch_configs ON app_branch_configs.id = public_git_vcs_configs.component_config_id AND public_git_vcs_configs.component_config_type = ?", appBranchConfigTable).
		Where("app_branch_configs.org_id = ?", req.OrgID).
		Where("(public_git_vcs_configs.repo = ? OR public_git_vcs_configs.repo = ?)", req.Repo, "https://github.com/"+req.Repo+".git").
		Where("public_git_vcs_configs.branch = ?", req.Branch).
		Where("public_git_vcs_configs.deleted_at = 0").
		Where("app_branch_configs.deleted_at = 0").
		Order("app_branch_configs.created_at DESC").
		Scan(&publicResults).Error
	if err != nil {
		return nil, fmt.Errorf("unable to find matching app branches (public): %w", err)
	}

	results = append(results, publicResults...)

	// Deduplicate: keep only the latest config per app branch
	seen := make(map[string]bool)
	var deduped []MatchingAppBranch
	for _, r := range results {
		if seen[r.AppBranchID] {
			continue
		}
		seen[r.AppBranchID] = true
		deduped = append(deduped, r)
	}

	return deduped, nil
}
