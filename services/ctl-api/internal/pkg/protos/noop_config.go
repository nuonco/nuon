package protos

import (
	"time"

	buildv1 "github.com/powertoolsdev/mono/pkg/types/components/build/v1"
	componentv1 "github.com/powertoolsdev/mono/pkg/types/components/component/v1"
	connectionsv1 "github.com/powertoolsdev/mono/pkg/types/components/connections/v1"
	deployv1 "github.com/powertoolsdev/mono/pkg/types/components/deploy/v1"
	"google.golang.org/protobuf/types/known/durationpb"
)

const (
	defaultBuildTimeout  time.Duration = time.Hour
	defaultDeployTimeout time.Duration = time.Minute * 5
)

func (c *Adapter) ToNoopConfig(componentID string) (*componentv1.Component, error) {
	return &componentv1.Component{
		Id: componentID,
		BuildCfg: &buildv1.Config{
			Timeout: durationpb.New(defaultBuildTimeout),
			Cfg: &buildv1.Config_Noop{
				Noop: &buildv1.NoopConfig{},
			},
		},
		DeployCfg: &deployv1.Config{
			Timeout: durationpb.New(defaultDeployTimeout),
			Cfg: &deployv1.Config_Noop{
				Noop: &deployv1.NoopConfig{},
			},
		},
		Connections: &connectionsv1.Connections{
			Instances: []*connectionsv1.InstanceConnection{},
		},
	}, nil
}
