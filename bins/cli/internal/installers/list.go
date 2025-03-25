package installers

import (
	"context"

	"github.com/nuonco/nuon-go/models"
	helpers "github.com/powertoolsdev/mono/bins/cli/internal"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) List(ctx context.Context, asJSON bool) error {
	view := ui.NewListView()

	installers, err := s.list(ctx)
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
	view.Render(data)
	return nil
}

func (s *Service) list(ctx context.Context) ([]*models.AppInstaller, error) {
	if !s.cfg.PaginationEnabled {
		installers, _, err := s.api.GetInstallers(ctx, &models.GetInstallersQuery{
			Offset:            0,
			Limit:             10,
			PaginationEnabled: s.cfg.PaginationEnabled,
		})
		if err != nil {
			return nil, err
		}
		return installers, nil
	}

	fetchFn := func(ctx context.Context, offset, limit int) ([]*models.AppInstaller, bool, error) {
		installers, hasMore, err := s.api.GetInstallers(ctx, &models.GetInstallersQuery{
			Offset:            offset,
			Limit:             limit,
			PaginationEnabled: s.cfg.PaginationEnabled,
		})
		if err != nil {
			return nil, false, err
		}
		return installers, hasMore, nil
	}

	return helpers.BatchFetch(ctx, 10, 50, fetchFn)
}
