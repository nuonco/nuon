package presets

import (
	buildv1 "github.com/powertoolsdev/protos/components/generated/types/build/v1"
	componentv1 "github.com/powertoolsdev/protos/components/generated/types/component/v1"
	deployv1 "github.com/powertoolsdev/protos/components/generated/types/deploy/v1"
	vcsv1 "github.com/powertoolsdev/protos/components/generated/types/vcs/v1"
)

func (p *preset) privateDockerHttpbin() (*componentv1.Component, error) {
	return &componentv1.Component{
		Id:   p.ID,
		Name: "private-docker-httpbin",
		BuildCfg: &buildv1.Config{
			Cfg: &buildv1.Config_DockerCfg{
				DockerCfg: &buildv1.DockerConfig{
					Dockerfile: "Dockerfile",
					VcsCfg: &vcsv1.Config{
						Cfg: &vcsv1.Config_PrivateGithubConfig{
							PrivateGithubConfig: &vcsv1.PrivateGithubConfig{
								Repo:      "jonmorehouse/go-httpbin",
								Directory: ".",
								// TODO(jm): add branch
								CommitRef:              "main",
								GithubAppKeyId:         "261597",
								GithubAppKeySecretName: "graphql-api-github-app-key",
								GithubInstallId:        "34504664",
							},
						},
					},
				},
			},
		},
		DeployCfg: &deployv1.Config{
			Cfg: &deployv1.Config_Basic{
				Basic: &deployv1.BasicConfig{
					InstanceCount: 1,
					ListenerCfg: &deployv1.ListenerConfig{
						ListenPort:      8080,
						HealthCheckPath: "/",
					},
				},
			},
		},
	}, nil
}
