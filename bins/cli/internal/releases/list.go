package releases

import (
	"context"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) List(ctx context.Context, appID, compID string, asJSON bool) {
	view := ui.NewListView()

	var (
		releases []*models.AppComponentRelease
		err      error
	)

	if appID != "" {
		releases, err = s.api.GetAppReleases(ctx, appID)
	} else if compID != "" {
		releases, err = s.api.GetComponentReleases(ctx, compID)
	}
	if err != nil {
		view.Error(err)
		return
	}

	if asJSON {
		ui.PrintJSON(releases)
		return
	}

	data := [][]string{
		{
			"id",
			"status",
			"build id",
			"created at",
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
}
