package components

import (
	"fmt"

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
				PublicGitConfig: &vcsv1.PublicGitConfig{
					Repo:      publicCfg.Repo,
					Directory: publicCfg.Directory,
					GitRef:    gitRef,
				},
			},
		}, nil
	}

	return &vcsv1.Config{
		Cfg: &vcsv1.Config_ConnectedGithubConfig{
			ConnectedGithubConfig: &vcsv1.ConnectedGithubConfig{
				Repo:                   ghCfg.Repo,
				Directory:              ghCfg.Directory,
				GitRef:                 gitRef,
				GithubAppKeyId:         c.cfg.GithubAppID,
				GithubAppKeySecretName: c.cfg.GithubAppKeySecretName,
				GithubInstallId:        ghCfg.VCSConnection.GithubInstallID,
			},
		},
	}, nil
}
