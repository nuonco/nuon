package sync

import (
	"context"

	"github.com/nuonco/nuon-go/models"
)

func (s sync) syncAppRunner(ctx context.Context, resource string) error {
	cfg, err := s.apiClient.CreateAppRunnerConfig(ctx, s.appID, &models.ServiceCreateAppRunnerConfigRequest{
		AppConfigID: s.appConfigID,
		EnvVars:     s.cfg.Runner.EnvVarMap,
		// TODO: after this PR is merged, update nuon-go, then enable this again
		// HelmDriver:  models.AppAppRunnerConfigHelmDriverType(s.cfg.Runner.HelmDriver),
		// Type:    models.AppAppRunnerType(s.cfg.Runner.RunnerType),
		Type: models.NewAppAppRunnerType(models.AppAppRunnerType(s.cfg.Runner.RunnerType)),
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
