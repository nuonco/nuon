package presets

import (
	buildv1 "github.com/powertoolsdev/protos/components/generated/types/build/v1"
	componentv1 "github.com/powertoolsdev/protos/components/generated/types/component/v1"
	deployv1 "github.com/powertoolsdev/protos/components/generated/types/deploy/v1"
	vcsv1 "github.com/powertoolsdev/protos/components/generated/types/vcs/v1"
)

func (p *preset) publicDockerHttpbin() (*componentv1.Component, error) {
	return &componentv1.Component{
		Id:   p.ID,
		Name: "public-docker-httpbin",
		BuildCfg: &buildv1.Config{
			Cfg: &buildv1.Config_DockerCfg{
				DockerCfg: &buildv1.DockerConfig{
					Dockerfile: "Dockerfile",
					VcsCfg: &vcsv1.Config{
						Cfg: &vcsv1.Config_PublicGithubConfig{
							PublicGithubConfig: &vcsv1.PublicGithubConfig{
								Repo:      "kennethreitz/httpbin",
								Directory: "/",
								Branch:    "main",
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
