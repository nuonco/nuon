package presets

import (
	buildv1 "github.com/powertoolsdev/mono/pkg/types/components/build/v1"
	componentv1 "github.com/powertoolsdev/mono/pkg/types/components/component/v1"
	deployv1 "github.com/powertoolsdev/mono/pkg/types/components/deploy/v1"
	"google.golang.org/protobuf/types/known/durationpb"
)

func (p *preset) privateExternalImageHttpbin() (*componentv1.Component, error) {
	return &componentv1.Component{
		Id: p.ID,
		BuildCfg: &buildv1.Config{
			Timeout: durationpb.New(defaultBuildTimeout),
			Cfg: &buildv1.Config_ExternalImageCfg{
				ExternalImageCfg: &buildv1.ExternalImageConfig{
					// NOTE: this is an internally built image in the sandbox-testing account for an
					// org
					OciImageUrl: "431927561584.dkr.ecr.us-west-2.amazonaws.com/demo/external-image-go-httpbin",
					Tag:         "v0.1.0",
					AuthCfg: &buildv1.ExternalImageAuthConfig{
						Cfg: &buildv1.ExternalImageAuthConfig_AwsIamAuthCfg{
							AwsIamAuthCfg: &buildv1.AWSIAMAuthCfg{
								IamRoleArn: "arn:aws:iam::949309607565:role/nuon-demo-external-image-access",
								AwsRegion:  "us-west-2",
							},
						},
					},
				},
			},
		},
		DeployCfg: &deployv1.Config{
			Timeout: durationpb.New(defaultDeployTimeout),
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
