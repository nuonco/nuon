package presets

import (
	buildv1 "github.com/powertoolsdev/mono/pkg/types/components/build/v1"
	componentv1 "github.com/powertoolsdev/mono/pkg/types/components/component/v1"
	deployv1 "github.com/powertoolsdev/mono/pkg/types/components/deploy/v1"
	vcsv1 "github.com/powertoolsdev/mono/pkg/types/components/vcs/v1"
)

func (p *preset) publicDockerHttpbin() (*componentv1.Component, error) {
	return &componentv1.Component{
		Id: p.ID,
		BuildCfg: &buildv1.Config{
			Cfg: &buildv1.Config_DockerCfg{
				DockerCfg: &buildv1.DockerConfig{
					Dockerfile: "Dockerfile",
					VcsCfg: &vcsv1.Config{
						Cfg: &vcsv1.Config_PublicGithubConfig{
							PublicGithubConfig: &vcsv1.PublicGithubConfig{
								Repo:      "https://github.com/postmanlabs/httpbin.git",
								Directory: ".",
								GitRef:    "master",
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
						ListenPort:      80,
						HealthCheckPath: "/",
					},
				},
			},
		},
	}, nil
}
