package installs

import (
	"context"
	"strconv"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/paginate"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	nuon "github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (s *Service) Components(ctx context.Context, installID string, offset, limit int, showAll, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}
	view := ui.NewListView()

	fetch := func(off, lim int) ([]*models.AppInstallComponent, bool, error) {
		return s.listComponents(ctx, installID, off, lim, showAll)
	}

	var (
		components []*models.AppInstallComponent
		hasMore    bool
	)
	if limit <= 0 {
		components, err = paginate.All(fetch)
	} else {
		components, hasMore, err = fetch(offset, limit)
	}
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
			"ENABLED",
			"HEALTH",
			"STATUS",
			"LATEST DEPLOY",
			"LATEST RELEASE",
		},
	}
	for _, comp := range components {
		enabled := "-"
		if comp.Enabled != nil {
			enabled = strconv.FormatBool(*comp.Enabled)
		}
		health := comp.HealthStatus
		if health == "" || health == "not-applicable" {
			health = "-"
		}
		args := []string{
			comp.Component.ID,
			comp.Component.Name,
			enabled,
			health,
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
	if limit <= 0 {
		view.RenderTotal(data, len(components))
	} else {
		view.RenderPaging(data, offset, limit, hasMore)
	}
	return nil
}

func (s *Service) listComponents(ctx context.Context, installID string, offset, limit int, showAll bool) ([]*models.AppInstallComponent, bool, error) {
	opts := nuon.GetInstallComponentsOpts{}
	if showAll {
		f := false
		opts.Synced = &f
	}
	cmps, hasMore, err := s.api.GetInstallComponents(ctx, installID, &models.GetPaginatedQuery{
		Offset: offset,
		Limit:  limit,
	}, opts)
	if err != nil {
		return nil, false, err
	}
	return cmps, hasMore, nil
}
