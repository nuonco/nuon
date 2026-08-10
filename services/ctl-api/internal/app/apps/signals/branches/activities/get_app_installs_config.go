package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetAppInstallsConfigOutput struct {
	ID              string  `json:"id"`
	VCSType         string  `json:"vcs_type"`
	VCSConnectionID *string `json:"vcs_connection_id,omitempty"`
	VCSConfigID     string  `json:"vcs_config_id,omitempty"`
	Repo            string  `json:"repo"`
	Branch          string  `json:"branch"`
	Directory       string  `json:"directory"`
	Source          string  `json:"source"`
	Found           bool    `json:"found"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
// @as-wrapper
// @by-field appID
func (a *Activities) getAppInstallsConfig(ctx context.Context, appID string) (*GetAppInstallsConfigOutput, error) {
	var cfg app.AppInstallsConfig
	err := a.db.WithContext(ctx).
		Preload("ConnectedGithubVCSConfig").
		Preload("PublicGitVCSConfig").
		Where(app.AppInstallsConfig{AppID: appID}).
		Order("created_at DESC").
		First(&cfg).Error
	if err != nil {
		return &GetAppInstallsConfigOutput{Found: false}, nil
	}

	var vcsConfigID string
	if cfg.ConnectedGithubVCSConfig != nil {
		vcsConfigID = cfg.ConnectedGithubVCSConfig.ID
	} else if cfg.PublicGitVCSConfig != nil {
		vcsConfigID = cfg.PublicGitVCSConfig.ID
	}

	return &GetAppInstallsConfigOutput{
		ID:              cfg.ID,
		VCSType:         cfg.VCSType,
		VCSConnectionID: cfg.VCSConnectionID,
		VCSConfigID:     vcsConfigID,
		Repo:            cfg.Repo,
		Branch:          cfg.Branch,
		Directory:       cfg.Directory,
		Source:          cfg.Source,
		Found:           true,
	}, nil
}
