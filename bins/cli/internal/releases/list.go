package releases

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/powertoolsdev/mono/pkg/api/client/models"
)

func (s *Service) List(ctx context.Context, appID, compID string) {
	view := ui.NewListView()

	releases := []*models.AppComponentRelease{}
	err := error(nil)
	if appID != "" {
		releases, err = s.api.GetAppReleases(ctx, appID)
	} else if compID != "" {
		releases, err = s.api.GetComponentReleases(ctx, compID)
	}
	if err != nil {
		view.Error(err)
		return
	}
	data := [][]string{
		[]string{
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
