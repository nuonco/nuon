package sync

import (
	"context"

	"github.com/nuonco/nuon-go/models"
)

func (s sync) syncAppRunner(ctx context.Context, resource string) error {
	newCfgEnvVars := make(map[string]string)
	for _, v := range s.cfg.Runner.EnvironmentVariables {
		newCfgEnvVars[v.Name] = v.Value
	}

	cfg, err := s.apiClient.CreateAppRunnerConfig(ctx, s.appID, &models.ServiceCreateAppRunnerConfigRequest{
		EnvVars: s.cfg.Runner.EnvVarMap,
		Type:    models.AppAppRunnerType(s.cfg.Runner.RunnerType),
	})
	if err != nil {
		return SyncAPIErr{
			Resource: resource,
			Err:      err,
		}
	}

	s.state.RunnerConfigID = cfg.ID
	return nil
}
