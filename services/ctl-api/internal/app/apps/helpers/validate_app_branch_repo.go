package helpers

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

// FetchAppBranchesWithConfigs retrieves all app branches for an app with their configs preloaded
func (h *Helpers) FetchAppBranchesWithConfigs(ctx context.Context, appID string) ([]app.AppBranch, error) {
	var branches []app.AppBranch

	// Load all branches for the app with their configs
	err := h.db.WithContext(ctx).
		Where("app_id = ?", appID).
		Preload("Configs", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		Preload("Configs.PublicGitVCSConfig").
		Preload("Configs.ConnectedGithubVCSConfig").
		Find(&branches).Error

	if err != nil {
		return nil, fmt.Errorf("unable to load app branches: %w", err)
	}

	return branches, nil
}

// ValidateSameRepo validates that all app branches use the same repository as the provided request
// Accepts the VCS config request (before creation) to validate early
// If both publicRepoReq and connectedRepoReq are nil, returns early with no validation
// Branches with no VCS config (both nil) are ignored during validation
func (h *Helpers) ValidateSameRepo(
	branches []app.AppBranch,
	vcsConfigReq *vcshelpers.VCSConfigRequest,
) error {
	// If no new repo is being added, skip validation
	if vcsConfigReq == nil {
		return nil
	}

	publicRepoReq := vcsConfigReq.PublicGitVCSConfig
	connectedRepoReq := vcsConfigReq.ConnectedGithubVCSConfig

	if publicRepoReq == nil && connectedRepoReq == nil {
		return nil
	}

	// Determine the new repo being added
	var newRepo string
	if publicRepoReq != nil {
		newRepo = publicRepoReq.Repo
	} else if connectedRepoReq != nil {
		newRepo = connectedRepoReq.Repo
	}

	// Check each branch's latest config
	for _, branch := range branches {
		// Skip branches without configs
		if len(branch.Configs) == 0 {
			continue
		}

		// Get the latest config (first one due to ORDER BY created_at DESC)
		latestConfig := branch.Configs[0]

		// Extract repo from the branch's latest config
		var branchRepo string

		if latestConfig.PublicGitVCSConfig != nil {
			branchRepo = latestConfig.PublicGitVCSConfig.Repo
		} else if latestConfig.ConnectedGithubVCSConfig != nil {
			branchRepo = latestConfig.ConnectedGithubVCSConfig.Repo
		} else {
			// Branch has no VCS config, ignore it
			continue
		}

		// Compare repos
		if branchRepo != newRepo {
			return stderr.ErrUser{
				Err: fmt.Errorf("repository mismatch across app branches"),
				Description: fmt.Sprintf(
					"all app branches must use the same repository. Branch '%s' uses '%s', but new config uses '%s'",
					branch.Name,
					branchRepo,
					newRepo,
				),
				Code: "app_branch_repo_mismatch",
			}
		}
	}

	return nil
}
