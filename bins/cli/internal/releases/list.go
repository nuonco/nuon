package releases

import (
	"context"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) List(ctx context.Context, appID, compID string, asJSON bool) error {
	view := ui.NewListView()

	var (
		releases []*models.AppComponentRelease
		err      error
	)

	if compID != "" {
		compID, err = lookup.ComponentID(ctx, s.api, appID, compID)
		if err != nil {
			return ui.PrintError(err)
		}

		releases, err = s.api.GetComponentReleases(ctx, compID)
	} else if appID != "" {
		appID, err = lookup.AppID(ctx, s.api, appID)
		if err != nil {
			return ui.PrintError(err)
		}

		releases, err = s.api.GetAppReleases(ctx, appID)
	}
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(releases)
		return nil
	}

	data := [][]string{
		{
			"ID",
			"STATUS",
			"BUILD ID",
			"CREATED AT",
		},
	}
	for _, release := range releases {
		data = append(data, []string{
			release.ID,
			release.Status,
			release.BuildID,
			release.CreatedAt,
		})
	}
	view.Render(data)
	return nil
}
