package installs

import (
	"context"

	"github.com/nuonco/nuon-go/models"
	helpers "github.com/powertoolsdev/mono/bins/cli/internal"
	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Components(ctx context.Context, installID string, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}
	view := ui.NewGetView()

	components, err := s.listComponents(ctx, installID)
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
	view.Render(data)
	return nil
}

func (s *Service) listComponents(ctx context.Context, installID string) ([]*models.AppInstallComponent, error) {
	if !s.cfg.PaginationEnabled {
		cmps, _, err := s.api.GetInstallComponents(ctx, installID, &models.GetInstallComponentsQuery{
			Offset:            0,
			Limit:             10,
			PaginationEnabled: s.cfg.PaginationEnabled,
		})
		if err != nil {
			return nil, err
		}
		return cmps, nil
	}

	fetchFn := func(ctx context.Context, offset, limit int) ([]*models.AppInstallComponent, bool, error) {
		cmps, hasMore, err := s.api.GetInstallComponents(ctx, installID, &models.GetInstallComponentsQuery{
			Offset:            offset,
			Limit:             limit,
			PaginationEnabled: s.cfg.PaginationEnabled,
		})
		if err != nil {
			return nil, false, err
		}
		return cmps, hasMore, nil
	}

	return helpers.BatchFetch(ctx, 10, 50, fetchFn)
}
