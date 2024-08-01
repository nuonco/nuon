package orgs

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) ConnectedRepos(ctx context.Context, asJSON bool) error {
	view := ui.NewGetView()

	repos, err := s.api.GetAllVCSConnectedRepos(ctx)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(repos)
		return nil
	}

	data := [][]string{
		{
			"USER",
			"NAME",
			"DEFAULT BRANCH",
			"GIT URL",
			"GITHUB INSTALL ID",
		},
	}

	for _, repo := range repos {
		data = append(data, []string{
			*repo.UserName,
			*repo.Name,
			*repo.DefaultBranch,
			*repo.GitURL,
			*repo.GithubInstallID,
		})
	}
	view.Render(data)
	return nil
}
