package sync

import (
	"context"

	"github.com/nuonco/nuon-go/models"
)

func (s sync) syncAppRunner(ctx context.Context, resource string) error {
	cfg, err := s.apiClient.CreateAppRunnerConfig(ctx, s.appID, &models.ServiceCreateAppRunnerConfigRequest{
		AppConfigID: s.appConfigID,
		EnvVars:     s.cfg.Runner.EnvVarMap,
		HelmDriver:  models.AppAppRunnerConfigHelmDriverType(s.cfg.Runner.HelmDriver),
		Type:        models.NewAppAppRunnerType(models.AppAppRunnerType(s.cfg.Runner.RunnerType)),
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
