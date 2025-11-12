package installs

import (
	"context"

	"github.com/nuonco/nuon-go/models"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) List(ctx context.Context, appID string, offset, limit int, asJSON bool) error {
	view := ui.NewListView()

	var (
		installs []*models.AppInstall
		hasMore  bool
		err      error
	)

	if appID != "" {
		appID, err := lookup.AppID(ctx, s.api, appID)
		if err != nil {
			return ui.PrintError(err)
		}
		installs, hasMore, err = s.listAppInstalls(ctx, appID, offset, limit)

	} else {
		installs, hasMore, err = s.listInstalls(ctx, offset, limit)
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
	view.RenderPaging(data, offset, limit, hasMore)
	return nil
}

func (s *Service) listInstalls(ctx context.Context, offset, limit int) ([]*models.AppInstall, bool, error) {
	installs, hasMore, err := s.api.GetAllInstalls(ctx, &models.GetPaginatedQuery{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return nil, hasMore, err
	}
	return installs, hasMore, nil
}

func (s *Service) listAppInstalls(ctx context.Context, appID string, offset, limit int) ([]*models.AppInstall, bool, error) {
	cmps, hasMore, err := s.api.GetAppInstalls(ctx, appID, &models.GetPaginatedQuery{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return nil, hasMore, err
	}
	return cmps, hasMore, nil
}
