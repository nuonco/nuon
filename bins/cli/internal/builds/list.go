package builds

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) List(ctx context.Context, compID string) {
	view := ui.NewBuildsListView()

	builds, err := s.api.GetComponentBuilds(ctx, compID)
	if err != nil {
		view.Error(err)
		return
	}

	view.Render(builds)
}
