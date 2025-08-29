package installs

import (
	"context"
	"time"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) SandboxRuns(ctx context.Context, installID string, offset, limit int, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	view := ui.NewListView()

	runs, hasMore, err := s.listSandboxRuns(ctx, installID, offset, limit)
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
	view.RenderPaging(data, offset, limit, hasMore)
	return nil
}

func (s *Service) listSandboxRuns(ctx context.Context, appID string, offset, limit int) ([]*models.AppInstallSandboxRun, bool, error) {
	runs, hasMore, err := s.api.GetInstallSandboxRuns(ctx, appID, &models.GetPaginatedQuery{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return nil, false, err
	}
	return runs, hasMore, nil
}
