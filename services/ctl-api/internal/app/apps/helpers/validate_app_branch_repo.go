package helpers

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

func (h *Helpers) FetchAppBranchesWithConfigs(ctx context.Context, appID string) ([]app.AppBranch, error) {
	var branches []app.AppBranch

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

func (h *Helpers) ValidateSameRepo(
	branches []app.AppBranch,
	vcsConfigReq *vcshelpers.VCSConfigRequest,
) error {
	if vcsConfigReq == nil {
		return nil
	}

	publicRepoReq := vcsConfigReq.PublicGitVCSConfig
	connectedRepoReq := vcsConfigReq.ConnectedGithubVCSConfig

	if publicRepoReq == nil && connectedRepoReq == nil {
		return nil
	}

	var newRepo string
	if publicRepoReq != nil {
		newRepo = publicRepoReq.Repo
	} else if connectedRepoReq != nil {
		newRepo = connectedRepoReq.Repo
	}

	for _, branch := range branches {
		if len(branch.Configs) == 0 {
			continue
		}

		latestConfig := branch.Configs[0]

		var branchRepo string

		if latestConfig.PublicGitVCSConfig != nil {
			branchRepo = latestConfig.PublicGitVCSConfig.Repo
		} else if latestConfig.ConnectedGithubVCSConfig != nil {
			branchRepo = latestConfig.ConnectedGithubVCSConfig.Repo
		} else {
			continue
		}

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
