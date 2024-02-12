package builds

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) List(ctx context.Context, compID, appID string, limit *int64, asJSON bool) {
	var err error
	if compID != "" {
		compID, err = lookup.ComponentID(ctx, s.api, compID)
		if err != nil {
			ui.PrintError(err)
			return
		}
	}
	if appID != "" {
		appID, err = lookup.AppID(ctx, s.api, appID)
		if err != nil {
			ui.PrintError(err)
			return
		}
	}

	view := ui.NewListView()

	builds, err := s.api.GetComponentBuilds(ctx, compID, appID, limit)
	if err != nil {
		view.Error(err)
		return
	}

	if asJSON {
		ui.PrintJSON(builds)
		return
	}

	data := [][]string{
		{
			"id",
			"status",
			"component name",
			"git ref / branch",
			"created at",
		},
	}
	for _, build := range builds {
		data = append(data, []string{
			build.ID,
			build.Status,
			build.ComponentName,
			build.GitRef,
			build.CreatedAt,
		})
	}
	view.Render(data)
}
