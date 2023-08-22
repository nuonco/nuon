package components

import (
	"github.com/powertoolsdev/mono/pkg/generics"
	deployv1 "github.com/powertoolsdev/mono/pkg/types/components/deploy/v1"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"google.golang.org/protobuf/types/known/durationpb"
)

func (c *Adapter) toBasicDeployConfig(syncOnly bool, cfg *app.BasicDeployConfig) *deployv1.Config {
	if syncOnly {
		return &deployv1.Config{
			Timeout: durationpb.New(defaultDeployTimeout),
			Cfg: &deployv1.Config_Noop{
				Noop: &deployv1.NoopConfig{},
			},
		}
	}

	return &deployv1.Config{
		Timeout: durationpb.New(defaultDeployTimeout),
		Cfg: &deployv1.Config_Basic{
			Basic: &deployv1.BasicConfig{
				InstanceCount: int32(cfg.InstanceCount),
				Args:          cfg.Args,
				ListenerCfg: &deployv1.ListenerConfig{
					ListenPort:      int32(cfg.ListenPort),
					HealthCheckPath: cfg.HealthCheckPath,
				},
				CpuRequest: generics.ToPtr(cfg.CPURequest),
				CpuLimit:   generics.ToPtr(cfg.CPULimit),
				MemRequest: generics.ToPtr(cfg.MemRequest),
				MemLimit:   generics.ToPtr(cfg.MemLimit),
				EnvVars:    c.toEnvVars(cfg.EnvVars),
			},
		},
	}
}
