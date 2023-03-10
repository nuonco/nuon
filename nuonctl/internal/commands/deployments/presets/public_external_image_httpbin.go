package presets

import (
	buildv1 "github.com/powertoolsdev/protos/components/generated/types/build/v1"
	componentv1 "github.com/powertoolsdev/protos/components/generated/types/component/v1"
	deployv1 "github.com/powertoolsdev/protos/components/generated/types/deploy/v1"
)

func (p *preset) publicExternalImageHttpbin() (*componentv1.Component, error) {
	return &componentv1.Component{
		Id:   p.ID,
		Name: "public-external-image-httpbin",
		BuildCfg: &buildv1.Config{
			Cfg: &buildv1.Config_ExternalImageCfg{
				ExternalImageCfg: &buildv1.ExternalImageConfig{
					OciImageUrl: "kennethreitz/httpbin",
					Tag:         "latest",
					AuthCfg: &buildv1.ExternalImageAuthConfig{
						Cfg: &buildv1.ExternalImageAuthConfig_PublicAuthCfg{},
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
