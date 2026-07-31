package apisyncer

import (
	"context"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

// runnerConfigRequest maps the parsed runner config onto the API request.
//
// Extracted from syncAppRunner so the mapping is testable without a nuon.Client: a
// field added to config.AppRunnerConfig and to the SDK model but never assigned here
// fails silently, which is exactly how phone_home_script_url went missing — the CLI
// accepted it, the API accepted it, and the value simply never left the machine.
// runner_sync_test.go pins every field against that.
func runnerConfigRequest(appConfigID string, cfg *config.AppRunnerConfig) *models.ServiceCreateAppRunnerConfigRequest {
	return &models.ServiceCreateAppRunnerConfigRequest{
		AppConfigID:        appConfigID,
		EnvVars:            cfg.EnvVarMap,
		HelmDriver:         models.AppAppRunnerConfigHelmDriverType(cfg.HelmDriver),
		InitScriptURL:      cfg.InitScriptURL,
		InstanceType:       cfg.InstanceType,
		RunnerAPIURL:       cfg.RunnerAPIURL,
		PublicAPIURL:       cfg.PublicAPIURL,
		PhoneHomeScriptURL: cfg.PhoneHomeScriptURL,
		Type:               models.NewAppAppRunnerType(models.AppAppRunnerType(cfg.RunnerType)),
	}
}

func (s *syncer) syncAppRunner(ctx context.Context, resource string) error {
	cfg, err := s.apiClient.CreateAppRunnerConfig(ctx, s.appID, runnerConfigRequest(s.appConfigID, s.cfg.Runner))
	if err != nil {
		return sync.SyncAPIErr{
			Resource: resource,
			Err:      err,
		}
	}

	s.state.RunnerConfigID = cfg.ID
	return nil
}
