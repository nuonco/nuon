package installs

import (
	"context"
	"time"

	"github.com/nuonco/nuon-go/models"
	helpers "github.com/powertoolsdev/mono/bins/cli/internal"
	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) SandboxRuns(ctx context.Context, installID string, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	view := ui.NewGetView()

	runs, err := s.listSandboxRuns(ctx, installID)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(runs)
		return nil
	}

	data := [][]string{
		{
			"ID",
			"RUN TYPE",
			"STATUS",
			"SANDBOX CONFIG TYPE",
			"SANDBOX REPO",
			"UPDATED AT",
		},
	}
	for _, run := range runs {
		var cfgType string
		var repo string

		if run.AppSandboxConfig.PublicGitVcsConfig != nil {
			cfgType = "public git"
			repo = run.AppSandboxConfig.PublicGitVcsConfig.Repo
		}

		if run.AppSandboxConfig.ConnectedGithubVcsConfig != nil {
			cfgType = "conntected github"
			repo = run.AppSandboxConfig.ConnectedGithubVcsConfig.Repo
		}

		updatedAt, _ := time.Parse(time.RFC3339Nano, run.UpdatedAt)

		data = append(data, []string{
			run.ID,
			string(run.RunType),
			run.StatusDescription,
			cfgType,
			repo,
			updatedAt.Format(time.Stamp),
		})
	}
	view.Render(data)
	return nil
}

func (s *Service) listSandboxRuns(ctx context.Context, appID string) ([]*models.AppInstallSandboxRun, error) {
	if !s.cfg.PaginationEnabled {
		runs, _, err := s.api.GetInstallSandboxRuns(ctx, appID, &models.GetInstallSandboxRunsQuery{
			Offset:            0,
			Limit:             10,
			PaginationEnabled: s.cfg.PaginationEnabled,
		})
		if err != nil {
			return nil, err
		}
		return runs, nil
	}

	fetchFn := func(ctx context.Context, offset, limit int) ([]*models.AppInstallSandboxRun, bool, error) {
		runs, hasMore, err := s.api.GetInstallSandboxRuns(ctx, appID, &models.GetInstallSandboxRunsQuery{
			Offset:            offset,
			Limit:             limit,
			PaginationEnabled: s.cfg.PaginationEnabled,
		})
		if err != nil {
			return nil, false, err
		}
		return runs, hasMore, nil
	}

	return helpers.BatchFetch(ctx, 10, 50, fetchFn)
}
