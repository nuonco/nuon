package buildv1

import (
	"reflect"

	"github.com/go-faker/faker/v4"
	vcsv1 "github.com/powertoolsdev/mono/pkg/types/components/vcs/v1"
)

func init() {
	_ = faker.AddProvider("buildConfig", fakeBuildConfig)
}

func fakeBuildConfig(v reflect.Value) (interface{}, error) {
	return &Config{
		Cfg: &Config_DockerCfg{
			DockerCfg: &DockerConfig{
				Dockerfile: "Dockerfile",
				VcsCfg: &vcsv1.Config{
					Cfg: &vcsv1.Config_PublicGitConfig{
						PublicGitConfig: &vcsv1.PublicGitConfig{
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
