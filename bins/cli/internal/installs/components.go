package installs

import (
	"context"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Components(ctx context.Context, installID string, offset, limit int, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}
	view := ui.NewListView()

	components, hasMore, err := s.listComponents(ctx, installID, offset, limit)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(components)
		return nil
	}

	data := [][]string{
		{
			"ID",
			"NAME",
			"STATUS",
			"LATEST DEPLOY",
			"LATEST RELEASE",
		},
	}
	for _, comp := range components {
		args := []string{
			comp.Component.ID,
			comp.Component.Name,
		}
		if len(comp.InstallDeploys) > 0 {
			args = append(args, []string{
				comp.InstallDeploys[0].Status,
				comp.InstallDeploys[0].ID,
				comp.InstallDeploys[0].ReleaseID,
			}...)
		} else {
			args = append(args, []string{
				"n/a",
				"n/a",
				"n/a",
			}...)
		}

		data = append(data, args)
	}
	view.RenderPaging(data, offset, limit, hasMore)
	return nil
}

func (s *Service) listComponents(ctx context.Context, installID string, offset, limit int) ([]*models.AppInstallComponent, bool, error) {
	cmps, hasMore, err := s.api.GetInstallComponents(ctx, installID, &models.GetPaginatedQuery{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return nil, false, err
	}
	return cmps, hasMore, nil
}
