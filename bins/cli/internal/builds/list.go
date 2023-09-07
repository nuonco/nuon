package builds

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) List(ctx context.Context, compID string) {
	view := ui.NewListView()

	builds, err := s.api.GetComponentBuilds(ctx, compID)
	if err != nil {
		view.Error(err)
		return
	}

	data := [][]string{
		[]string{
			"id",
			"status",
			"component id",
			"git ref",
		},
	}
	for _, build := range builds {
		data = append(data, []string{
			build.ID,
			build.Status,
			build.ComponentConfigConnectionID,
			build.GitRef,
		})
	}
	view.Render(data)
}
