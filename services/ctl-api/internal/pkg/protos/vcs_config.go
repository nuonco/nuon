package protos

import (
	"fmt"
	"strings"

	vcsv1 "github.com/powertoolsdev/mono/pkg/types/components/vcs/v1"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (c *Adapter) ToVCSConfig(gitRef string, publicCfg *app.PublicGitVCSConfig, ghCfg *app.ConnectedGithubVCSConfig) (*vcsv1.Config, error) {
	if publicCfg == nil && ghCfg == nil {
		return nil, fmt.Errorf("both public and connected github configs are nil")
	}

	if publicCfg != nil {
		return &vcsv1.Config{
			Cfg: &vcsv1.Config_PublicGitConfig{
				PublicGitConfig: c.toPublicGitConfig(publicCfg.Branch, publicCfg),
			},
		}, nil
	}

	return &vcsv1.Config{
		Cfg: &vcsv1.Config_ConnectedGithubConfig{
			ConnectedGithubConfig: c.toConnectedGithubConfig(gitRef, ghCfg),
		},
	}, nil
}

func (c *Adapter) toPublicGitConfig(gitRef string, cfg *app.PublicGitVCSConfig) *vcsv1.PublicGitConfig {
	repo := cfg.Repo
	if !strings.HasPrefix(repo, "https://") {
		repo = fmt.Sprintf("https://github.com/%s", repo)
	}

	return &vcsv1.PublicGitConfig{
		Repo:      repo,
		Directory: cfg.Directory,
		GitRef:    gitRef,
	}
}

func (c *Adapter) toConnectedGithubConfig(gitRef string, cfg *app.ConnectedGithubVCSConfig) *vcsv1.ConnectedGithubConfig {
	return &vcsv1.ConnectedGithubConfig{
		Repo:                   cfg.Repo,
		Directory:              cfg.Directory,
		GitRef:                 gitRef,
		GithubAppKeyId:         c.cfg.GithubAppID,
		GithubAppKeySecretName: c.cfg.GithubAppKeySecretName,
		GithubInstallId:        cfg.VCSConnection.GithubInstallID,
	}
}
