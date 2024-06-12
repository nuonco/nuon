package installs

import (
	"context"
	"time"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) SandboxRuns(ctx context.Context, installID string, asJSON bool) {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		ui.PrintError(err)
		return
	}

	view := ui.NewGetView()

	runs, err := s.api.GetInstallSandboxRuns(ctx, installID)
	if err != nil {
		view.Error(err)
		return
	}

	if asJSON {
		ui.PrintJSON(runs)
		return
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
}
