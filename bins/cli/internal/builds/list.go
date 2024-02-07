package builds

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) List(ctx context.Context, compID string, asJSON bool) {
	compID, err := lookup.ComponentID(ctx, s.api, compID)
	if err != nil {
		ui.PrintError(err)
		return
	}

	view := ui.NewListView()

	builds, err := s.api.GetComponentBuilds(ctx, compID)
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
			"component id",
			"git ref / branch",
		},
	}
	for _, build := range builds {
		data = append(data, []string{
			build.ID,
			build.Status,
			compID,
			build.GitRef,
		})
	}
	view.Render(data)
}
