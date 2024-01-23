package apps

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) GetRunnerConfig(ctx context.Context, appID string, asJSON bool) {
	appID, err := lookup.AppID(ctx, s.api, appID)
	if err != nil {
		ui.PrintError(err)
		return
	}

	view := ui.NewGetView()

	runnerCfg, err := s.api.GetAppRunnerLatestConfig(ctx, appID)
	if err != nil {
		view.Error(err)
		return
	}

	if asJSON {
		ui.PrintJSON(runnerCfg)
		return
	}

	args := [][]string{
		{"app_id", string(runnerCfg.AppID)},
		{"type", string(runnerCfg.AppRunnerType)},
	}
	for k, v := range runnerCfg.EnvVars {
		args = append(args, []string{
			"env-var", fmt.Sprintf("%s=%s", k, v),
		})
	}

	args = append(args, [][]string{
		{"created at", runnerCfg.CreatedAt},
		{"updated at", runnerCfg.UpdatedAt},
		{"created by", runnerCfg.CreatedByID},
	}...)
	view.Render(args)
}
