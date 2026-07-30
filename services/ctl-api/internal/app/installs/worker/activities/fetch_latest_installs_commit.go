package activities

import (
	"context"
	"encoding/json"
	"fmt"

	pkgconfig "github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
)

type FetchLatestInstallsCommitOutput struct {
	CommitID string `json:"commit_id"`
	SHA      string `json:"sha"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
// @as-wrapper
// @by-field appID
func (a *Activities) fetchLatestInstallsCommit(ctx context.Context, appID string) (*FetchLatestInstallsCommitOutput, error) {
	var latestConfig app.AppConfig
	if err := a.db.WithContext(ctx).
		Where(app.AppConfig{AppID: appID}).
		Order("created_at DESC").
		First(&latestConfig).Error; err != nil {
		return nil, fmt.Errorf("unable to get latest app config: %w", err)
	}

	if latestConfig.IntermediateConfig == nil || !latestConfig.IntermediateConfig.IsSet() {
		return nil, nil
	}

	blobCtx := blobstore.WithBlobService(ctx, a.blobSvc)
	intermediateJSON, err := latestConfig.IntermediateConfig.Get(blobCtx)
	if err != nil || intermediateJSON == "" {
		return nil, nil
	}

	var parsed pkgconfig.AppConfig
	if err := json.Unmarshal([]byte(intermediateJSON), &parsed); err != nil {
		return nil, nil
	}

	if parsed.InstallsConfig == nil {
		return nil, nil
	}

	ic := parsed.InstallsConfig

	var parentApp app.App
	if err := a.db.WithContext(ctx).
		Preload("Org").
		Preload("Org.VCSConnections").
		First(&parentApp, "id = ?", appID).Error; err != nil {
		return nil, fmt.Errorf("unable to get app: %w", err)
	}

	if ic.ConnectedRepo != nil {
		var vcsCfg app.ConnectedGithubVCSConfig
		if err := a.db.WithContext(ctx).
			Preload("VCSConnection").
			Where("repo = ?", ic.ConnectedRepo.Repo).
			Where("org_id = ?", parentApp.OrgID).
			Order("created_at DESC").
			First(&vcsCfg).Error; err != nil {
			return nil, fmt.Errorf("unable to find VCS config for repo %s: %w", ic.ConnectedRepo.Repo, err)
		}

		ghCommit, err := a.vcsHelpers.GetConnectedGithubVCSConfigLatestCommit(ctx, &vcsCfg)
		if err != nil {
			return nil, fmt.Errorf("unable to fetch latest commit: %w", err)
		}

		vcsCommit := a.vcsHelpers.GithubCommitToVCSConnectionCommit(ghCommit, vcsCfg.ID, "ConnectedGithubVCSConfig", vcsCfg.VCSConnectionID)
		if vcsCommit == nil {
			return nil, nil
		}

		if err := a.db.WithContext(ctx).Create(vcsCommit).Error; err != nil {
			return nil, fmt.Errorf("unable to create commit record: %w", err)
		}

		return &FetchLatestInstallsCommitOutput{CommitID: vcsCommit.ID, SHA: vcsCommit.SHA}, nil
	}

	if ic.PublicRepo != nil {
		var publicCfg app.PublicGitVCSConfig
		publicCfg.Repo = ic.PublicRepo.Repo
		publicCfg.Branch = ic.PublicRepo.Branch

		ghCommit, err := a.vcsHelpers.GetPublicGitVCSConfigLatestCommit(ctx, &publicCfg)
		if err != nil {
			return nil, fmt.Errorf("unable to fetch latest commit: %w", err)
		}

		vcsCommit := a.vcsHelpers.GithubCommitToVCSConnectionCommit(ghCommit, "", "PublicGitVCSConfig", "")
		if vcsCommit == nil {
			return nil, nil
		}

		if err := a.db.WithContext(ctx).Create(vcsCommit).Error; err != nil {
			return nil, fmt.Errorf("unable to create commit record: %w", err)
		}

		return &FetchLatestInstallsCommitOutput{CommitID: vcsCommit.ID, SHA: vcsCommit.SHA}, nil
	}

	return nil, nil
}
