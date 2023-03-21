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
					OciImageUrl: "some-private-ecr-repo",
					Tag:         "latest",
					AuthCfg: &buildv1.ExternalImageAuthConfig{
						Cfg: &buildv1.ExternalImageAuthConfig_AwsIamAuthCfg{
							AwsIamAuthCfg: &buildv1.AWSIAMAuthCfg{
								IamRoleArn: "some-iam-role-arn",
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
