package protos

import (
	buildv1 "github.com/powertoolsdev/mono/pkg/types/components/build/v1"
	componentv1 "github.com/powertoolsdev/mono/pkg/types/components/component/v1"
	deployv1 "github.com/powertoolsdev/mono/pkg/types/components/deploy/v1"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"google.golang.org/protobuf/types/known/durationpb"
)

func (c *Adapter) toExternalImageAuthConfig(cfg *app.ExternalImageComponentConfig) *buildv1.ExternalImageAuthConfig {
	if cfg.AWSECRImageConfig == nil {
		return &buildv1.ExternalImageAuthConfig{
			Cfg: &buildv1.ExternalImageAuthConfig_PublicAuthCfg{
				PublicAuthCfg: &buildv1.PublicAuthCfg{},
			},
		}
	}

	return &buildv1.ExternalImageAuthConfig{
		Cfg: &buildv1.ExternalImageAuthConfig_AwsIamAuthCfg{
			AwsIamAuthCfg: &buildv1.AWSIAMAuthCfg{
				IamRoleArn: cfg.AWSECRImageConfig.IAMRoleARN,
				AwsRegion:  cfg.AWSECRImageConfig.AWSRegion,
			},
		},
	}
}

func (c *Adapter) ToExternalImageConfig(cfg *app.ExternalImageComponentConfig, connections []app.InstallDeploy) (*componentv1.Component, error) {
	return &componentv1.Component{
		Id: cfg.ComponentConfigConnection.ComponentID,
		BuildCfg: &buildv1.Config{
			Timeout: durationpb.New(defaultBuildTimeout),
			Cfg: &buildv1.Config_ExternalImageCfg{
				ExternalImageCfg: &buildv1.ExternalImageConfig{
					Tag:         cfg.Tag,
					OciImageUrl: cfg.ImageURL,
					AuthCfg:     c.toExternalImageAuthConfig(cfg),
				},
			},
		},
		DeployCfg: &deployv1.Config{
			Timeout: durationpb.New(defaultDeployTimeout),
			Cfg: &deployv1.Config_Noop{
				Noop: &deployv1.NoopConfig{},
			},
		},
		Connections: c.toConnections(connections),
	}, nil
}
