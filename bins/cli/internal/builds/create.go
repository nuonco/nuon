package builds

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/powertoolsdev/mono/pkg/api/client/models"
)

func (s *Service) Create(ctx context.Context, compID string) {
	view := ui.NewCreateView("build")

	view.Start()
	build, err := s.api.CreateComponentBuild(
		ctx,
		compID,
		&models.ServiceCreateComponentBuildRequest{
			UseLatest: true,
		},
	)
	if err != nil {
		view.Fail(err)
		return
	}

	view.Success(build.ID)
}
