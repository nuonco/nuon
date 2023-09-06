package builds

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Get(ctx context.Context, compID, buildID string) {
	view := ui.NewBuildsGetView()

	build, err := s.api.GetComponentBuild(ctx, compID, buildID)
	if err != nil {
		view.Error(err)
		return
	}

	view.Render(build)
}
