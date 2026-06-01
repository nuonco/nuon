package apisyncer

import (
	"context"

	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (s *syncer) syncAppRunner(ctx context.Context, resource string) error {
	cfg, err := s.apiClient.CreateAppRunnerConfig(ctx, s.appID, &models.ServiceCreateAppRunnerConfigRequest{
		AppConfigID:   s.appConfigID,
		EnvVars:       s.cfg.Runner.EnvVarMap,
		HelmDriver:    models.AppAppRunnerConfigHelmDriverType(s.cfg.Runner.HelmDriver),
		InitScriptURL: s.cfg.Runner.InitScriptURL,
		InstanceType:  s.cfg.Runner.InstanceType,
		Type:          models.NewAppAppRunnerType(models.AppAppRunnerType(s.cfg.Runner.RunnerType)),
	})
	if err != nil {
		return sync.SyncAPIErr{
			Resource: resource,
			Err:      err,
		}
	}

	s.state.RunnerConfigID = cfg.ID
	return nil
}
