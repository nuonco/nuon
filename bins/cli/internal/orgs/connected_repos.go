package orgs

import (
	"context"

	"github.com/nuonco/nuon-go/models"
	helpers "github.com/powertoolsdev/mono/bins/cli/internal"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) ConnectedRepos(ctx context.Context, asJSON bool) error {
	view := ui.NewGetView()

	repos, err := s.listConnectedRepos(ctx)
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

func (s *Service) listConnectedRepos(ctx context.Context) ([]*models.ServiceRepository, error) {
	if !s.cfg.PaginationEnabled {
		repos, _, err := s.api.GetAllVCSConnectedRepos(ctx, &models.GetAllVCSConnectedReposQuery{
			Offset:            0,
			Limit:             10,
			PaginationEnabled: s.cfg.PaginationEnabled,
		})
		if err != nil {
			return nil, err
		}
		return repos, nil
	}

	fetchFn := func(ctx context.Context, offset, limit int) ([]*models.ServiceRepository, bool, error) {
		repos, hasMore, err := s.api.GetAllVCSConnectedRepos(ctx, &models.GetAllVCSConnectedReposQuery{
			Offset:            offset,
			Limit:             limit,
			PaginationEnabled: s.cfg.PaginationEnabled,
		})
		if err != nil {
			return nil, false, err
		}
		return repos, hasMore, nil
	}

	return helpers.BatchFetch(ctx, 10, 50, fetchFn)
}
