package protos

import (
	buildv1 "github.com/powertoolsdev/mono/pkg/types/components/build/v1"
	componentv1 "github.com/powertoolsdev/mono/pkg/types/components/component/v1"
	deployv1 "github.com/powertoolsdev/mono/pkg/types/components/deploy/v1"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"google.golang.org/protobuf/types/known/durationpb"
)

func (c *Adapter) ToJobConfig(cfg *app.JobComponentConfig, connections []app.InstallDeploy) (*componentv1.Component, error) {
	return &componentv1.Component{
		Id: cfg.ComponentConfigConnection.ComponentID,
		BuildCfg: &buildv1.Config{
			Timeout: durationpb.New(defaultBuildTimeout),
			Cfg: &buildv1.Config_Noop{
				Noop: &buildv1.NoopConfig{},
			},
		},
		DeployCfg: &deployv1.Config{
			Timeout: durationpb.New(defaultDeployTimeout),
			Cfg: &deployv1.Config_Job{
				Job: &deployv1.JobConfig{
					Tag:      cfg.Tag,
					ImageUrl: cfg.ImageURL,
					Cmd:      cfg.Cmd,
					EnvVars:  c.toEnvVars(cfg.EnvVars),
					Args:     cfg.Args,
				},
			},
		},
		Connections: c.toConnections(connections),
	}, nil
}
