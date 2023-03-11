package fakers

import (
	"reflect"

	buildv1 "github.com/powertoolsdev/mono/pkg/protos/components/generated/types/build/v1"
	vcsv1 "github.com/powertoolsdev/mono/pkg/protos/components/generated/types/vcs/v1"
)

func fakeBuildConfig(v reflect.Value) (interface{}, error) {
	return &buildv1.Config{
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
	}, nil
}
