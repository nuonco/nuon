package orgs

import (
	"context"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) ConnectedRepos(ctx context.Context, offset, limit int, asJSON bool) error {
	view := ui.NewListView()

	repos, hasMore, err := s.listConnectedRepos(ctx, offset, limit)
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
	view.RenderPaging(data, offset, limit, hasMore)
	return nil
}

func (s *Service) listConnectedRepos(ctx context.Context, offset, limit int) ([]*models.ServiceRepository, bool, error) {
	repos, hasMore, err := s.api.GetAllVCSConnectedRepos(ctx, &models.GetPaginatedQuery{
		Offset:            offset,
		Limit:             limit,
		PaginationEnabled: true,
	})
	if err != nil {
		return nil, false, err
	}
	return repos, hasMore, nil
}
