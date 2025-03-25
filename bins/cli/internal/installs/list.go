package installs

import (
	"context"

	"github.com/nuonco/nuon-go/models"

	helpers "github.com/powertoolsdev/mono/bins/cli/internal"
	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) List(ctx context.Context, appID string, asJSON bool) error {
	view := ui.NewListView()

	var (
		installs []*models.AppInstall
		err      error
	)

	if appID != "" {
		appID, err := lookup.AppID(ctx, s.api, appID)
		if err != nil {
			return ui.PrintError(err)
		}
		installs, err = s.listAppInstalls(ctx, appID)

	} else {
		installs, err = s.listInstalls(ctx)
	}
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(installs)
		return nil
	}

	data := [][]string{
		{
			"NAME",
			"ID",
			"SANDBOX",
			"RUNNER",
			"COMPONENTS",
			"CREATED AT",
		},
	}
	curID := s.cfg.GetString("org_id")
	for _, install := range installs {
		if curID != "" {
			if install.ID == curID {
				install.Name = "*" + install.Name
			} else {
				install.Name = " " + install.Name
			}
		}
		data = append(data, []string{
			install.Name,
			install.ID,
			install.SandboxStatus,
			install.RunnerStatus,
			install.CompositeComponentStatus,
			install.CreatedAt,
		})
	}
	view.Render(data)
	return nil
}

func (s *Service) listInstalls(ctx context.Context) ([]*models.AppInstall, error) {
	if !s.cfg.PaginationEnabled {
		installs, _, err := s.api.GetAllInstalls(ctx, &models.GetAllInstallsQuery{
			Offset:            0,
			Limit:             10,
			PaginationEnabled: s.cfg.PaginationEnabled,
		})
		if err != nil {
			return nil, err
		}
		return installs, nil
	}

	fetchFn := func(ctx context.Context, offset, limit int) ([]*models.AppInstall, bool, error) {
		installs, hasMore, err := s.api.GetAllInstalls(ctx, &models.GetAllInstallsQuery{
			Offset:            offset,
			Limit:             limit,
			PaginationEnabled: s.cfg.PaginationEnabled,
		})
		if err != nil {
			return nil, false, err
		}
		return installs, hasMore, nil
	}

	return helpers.BatchFetch(ctx, 10, 50, fetchFn)
}

func (s *Service) listAppInstalls(ctx context.Context, appID string) ([]*models.AppInstall, error) {
	if !s.cfg.PaginationEnabled {
		cmps, _, err := s.api.GetAppInstalls(ctx, appID, &models.GetAppInstallsQuery{
			Offset:            0,
			Limit:             10,
			PaginationEnabled: s.cfg.PaginationEnabled,
		})
		if err != nil {
			return nil, err
		}
		return cmps, nil
	}

	fetchFn := func(ctx context.Context, offset, limit int) ([]*models.AppInstall, bool, error) {
		installs, hasMore, err := s.api.GetAppInstalls(ctx, appID, &models.GetAppInstallsQuery{
			Offset:            offset,
			Limit:             limit,
			PaginationEnabled: s.cfg.PaginationEnabled,
		})
		if err != nil {
			return nil, false, err
		}
		return installs, hasMore, nil
	}

	return helpers.BatchFetch(ctx, 10, 50, fetchFn)
}
