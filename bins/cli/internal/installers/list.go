package installers

import (
	"context"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) List(ctx context.Context, offset, limit int, asJSON bool) error {
	view := ui.NewListView()

	installers, hasMore, err := s.list(ctx, offset, limit)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(installers)
		return nil
	}

	data := [][]string{
		{
			"ID",
			"NAME",
			"CREATED AT",
		},
	}
	for _, installer := range installers {
		data = append(data, []string{
			installer.ID,
			installer.Metadata.Name,
			installer.CreatedAt,
		})
	}
	view.RenderPaging(data, offset, limit, hasMore)
	return nil
}

func (s *Service) list(ctx context.Context, offset, limit int) ([]*models.AppInstaller, bool, error) {
	installers, hasMore, err := s.api.GetInstallers(ctx, &models.GetPaginatedQuery{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return nil, hasMore, err
	}
	return installers, hasMore, nil
}
