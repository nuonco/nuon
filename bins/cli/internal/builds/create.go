package builds

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/powertoolsdev/mono/pkg/api/client/models"
)

func (s *Service) Create(ctx context.Context, compID string) {
	view := ui.NewUpdateView()

	view.Update(fmt.Sprintf("Starting build for component %s", compID))
	build, err := s.api.CreateComponentBuild(
		ctx,
		compID,
		&models.ServiceCreateComponentBuildRequest{
			UseLatest: true,
		},
	)
	if err != nil {
		view.Fail(fmt.Sprintf("build failed: %s", err))
		return
	}

	view.Success(fmt.Sprintf("build completed: %s", build.ID))
}
